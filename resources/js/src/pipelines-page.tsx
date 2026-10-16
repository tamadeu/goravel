import { type FormEvent, useDeferredValue, useEffect, useState } from "react"
import { ArrowLeft, ArrowRight, Check, DatabaseZap, LoaderCircle, MoreHorizontal, Plug, Plus, RefreshCw, Search, TriangleAlert, Workflow } from "lucide-react"

import { authTokenKey } from "@/auth-page"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"

type Connection = { id: number; name: string; kind: string; status: string }
type Pipeline = { id: number; name: string; description: string; source: Connection; destination: Connection; sourceSchema: string; sourceTable: string; destinationTable: string; columns: string[]; frequency: string }
type PipelineInput = { name: string; description: string; sourceId: number; destinationId: number; sourceSchema: string; sourceTable: string; destinationTable: string; columns: string[]; frequency: string }
type SourceTable = { schema: string; name: string }
type SourceColumn = { name: string; type: string; nullable: boolean }
type Metadata<T> = { key: string; status: "loading" | "ready" | "error"; data: T[]; error: string }

const frequencies = [{ value: "manual", label: "Manual" }, { value: "every_5_minutes", label: "A cada 5 minutos" }, { value: "every_15_minutes", label: "A cada 15 minutos" }, { value: "hourly", label: "A cada hora" }, { value: "daily", label: "Diariamente" }]
const wizardSteps = ["Destino", "Conector / origem", "Tabela da origem", "Tabela nova no destino", "Colunas da origem", "Nome e frequência"]
const supportsDiscovery = (connection: Connection) => connection.kind === "postgresql" || connection.kind === "clickhouse"
const tableKey = (table: SourceTable) => JSON.stringify([table.schema, table.name])
const suggestDestinationTable = (table: string) => table.replace(/[^A-Za-z0-9_]/g, "_").replace(/^(?=[0-9])/, "_").slice(0, 63)

const technologyNames: Record<string, string> = { postgresql: "PostgreSQL", mysql: "MySQL", bigquery: "Google BigQuery", clickhouse: "ClickHouse", sqlserver: "SQL Server", s3: "Amazon S3", rest: "REST API" }
const inputClass = "h-10 w-full min-w-0 rounded-md border border-border bg-background px-3 text-sm font-normal outline-none focus:ring-2 focus:ring-ring"

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, { ...init, headers: { "Content-Type": "application/json", Authorization: `Bearer ${localStorage.getItem(authTokenKey)}`, ...init?.headers } })
  if (response.status === 401) {
    localStorage.removeItem(authTokenKey)
    window.location.reload()
    throw new Error("Sua sessão expirou")
  }
  if (response.status === 204) return undefined as T
  const payload = await response.json() as { data?: T; message?: string }
  if (!response.ok) throw new Error(payload.message || "Não foi possível acessar os pipelines")
  return payload.data as T
}

function ConnectionSummary({ connection, destination = false }: { connection: Connection; destination?: boolean }) {
  const Icon = destination ? DatabaseZap : Plug
  return <div className="flex min-w-0 items-center gap-3">
    <span className={`grid size-9 shrink-0 place-items-center rounded-md ${destination ? "bg-[#e8f0fe] text-[#2458a6]" : "bg-emerald-50 text-emerald-700"}`}><Icon className="size-4" /></span>
    <div className="min-w-0"><p className="truncate text-sm font-semibold" title={connection.name}>{connection.name}</p><p className="truncate text-xs text-muted-foreground">{technologyNames[connection.kind] || connection.kind}</p></div>
  </div>
}

function PipelineDialog({ open, pipeline, sources, destinations, onOpenChange, onSave }: { open: boolean; pipeline: Pipeline | null; sources: Connection[]; destinations: Connection[]; onOpenChange: (open: boolean) => void; onSave: (input: PipelineInput) => Promise<void> }) {
  const [name, setName] = useState("")
  const [sourceId, setSourceId] = useState("")
  const [destinationId, setDestinationId] = useState("")
  const [sourceSchema, setSourceSchema] = useState("")
  const [sourceTable, setSourceTable] = useState("")
  const [destinationTable, setDestinationTable] = useState("")
  const [destinationDerived, setDestinationDerived] = useState(true)
  const [columns, setColumns] = useState<string[]>([])
  const [frequency, setFrequency] = useState("manual")
  const [step, setStep] = useState(0)
  const [tables, setTables] = useState<Metadata<SourceTable> | null>(null)
  const [columnMetadata, setColumnMetadata] = useState<Metadata<SourceColumn> | null>(null)
  const [tablesRetry, setTablesRetry] = useState(0)
  const [columnsRetry, setColumnsRetry] = useState(0)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState("")

  useEffect(() => {
    if (!open) return
    setName(pipeline?.name || "")
    setSourceId(pipeline?.source.id.toString() || "")
    setDestinationId(pipeline?.destination.id.toString() || "")
    setSourceSchema(pipeline?.sourceSchema || "")
    setSourceTable(pipeline?.sourceTable || "")
    setDestinationTable(pipeline?.destinationTable || "")
    setDestinationDerived(!pipeline?.destinationTable)
    setColumns(Array.from(new Set(pipeline?.columns || [])))
    setFrequency(pipeline?.frequency || "manual")
    setStep(0)
    setTables(null)
    setColumnMetadata(null)
    setError("")
  }, [open, pipeline])

  const source = sources.find((item) => item.id.toString() === sourceId)
  const destination = destinations.find((item) => item.id.toString() === destinationId)
  const discoverable = !!source && supportsDiscovery(source)
  const selectedTableKey = sourceTable ? tableKey({ schema: sourceSchema, name: sourceTable }) : ""
  const columnsKey = JSON.stringify([sourceId, sourceSchema, sourceTable])
  const tablesReady = tables?.key === sourceId && tables.status === "ready"
  const availableTables = tablesReady ? tables.data : []
  const tableValid = tablesReady && availableTables.some((table) => tableKey(table) === selectedTableKey)
  const columnsReady = columnMetadata?.key === columnsKey && columnMetadata.status === "ready"
  const availableColumns = columnsReady ? columnMetadata.data : []
  const unavailableColumns = columnsReady ? columns.filter((column) => !availableColumns.some((item) => item.name === column)) : []
  const destinationTableValid = /^[A-Za-z_][A-Za-z0-9_]*$/.test(destinationTable) && destinationTable.length <= 63
  const validity = [!!destination, discoverable, discoverable && tableValid, destinationTableValid, columnsReady && columns.length > 0 && unavailableColumns.length === 0, name.trim().length >= 2 && name.trim().length <= 120 && frequencies.some((item) => item.value === frequency)]
  const allValid = validity.every(Boolean)
  const allColumnsSelected = availableColumns.length > 0 && availableColumns.every((column) => columns.includes(column.name))

  useEffect(() => {
    if (!open || !discoverable) return
    const controller = new AbortController()
    setTables({ key: sourceId, status: "loading", data: [], error: "" })
    void request<SourceTable[]>(`/api/connectors/${sourceId}/tables`, { signal: controller.signal }).then((data) => {
      if (!controller.signal.aborted) setTables({ key: sourceId, status: "ready", data, error: "" })
    }).catch((requestError: unknown) => {
      if (!controller.signal.aborted) setTables({ key: sourceId, status: "error", data: [], error: requestError instanceof Error ? requestError.message : "Não foi possível carregar as tabelas" })
    })
    return () => controller.abort()
  }, [open, discoverable, sourceId, tablesRetry])

  useEffect(() => {
    if (!open || !discoverable || !tableValid) return
    const controller = new AbortController()
    setColumnMetadata({ key: columnsKey, status: "loading", data: [], error: "" })
    const params = new URLSearchParams({ schema: sourceSchema, table: sourceTable })
    void request<SourceColumn[]>(`/api/connectors/${sourceId}/columns?${params}`, { signal: controller.signal }).then((data) => {
      if (!controller.signal.aborted) setColumnMetadata({ key: columnsKey, status: "ready", data, error: "" })
    }).catch((requestError: unknown) => {
      if (!controller.signal.aborted) setColumnMetadata({ key: columnsKey, status: "error", data: [], error: requestError instanceof Error ? requestError.message : "Não foi possível carregar as colunas" })
    })
    return () => controller.abort()
  }, [open, discoverable, tableValid, columnsKey, sourceId, sourceSchema, sourceTable, columnsRetry])

  function changeSource(next: string) {
    if (next === sourceId) return
    setSourceId(next)
    setSourceSchema("")
    setSourceTable("")
    setColumns([])
    setDestinationTable("")
    setDestinationDerived(true)
    setTables(null)
    setColumnMetadata(null)
    setError("")
  }

  function changeTable(next: string) {
    const table = availableTables.find((item) => tableKey(item) === next)
    if (!table || next === selectedTableKey) return
    setSourceSchema(table.schema)
    setSourceTable(table.name)
    setColumns([])
    setColumnMetadata(null)
    if (destinationDerived) setDestinationTable(suggestDestinationTable(table.name))
    setError("")
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (saving) return
    setError("")
    if (step < 5) {
      if (validity[step]) setStep((current) => current + 1)
      else setError("Preencha os dados válidos desta etapa para continuar.")
      return
    }
    if (!allValid) {
      setStep(validity.findIndex((valid) => !valid))
      setError("Revise os dados desta etapa antes de salvar.")
      return
    }
    setSaving(true)
    try {
      await onSave({ name: name.trim(), description: pipeline?.description || "", sourceId: Number(sourceId), destinationId: Number(destinationId), sourceSchema, sourceTable, destinationTable, columns, frequency })
      onOpenChange(false)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Não foi possível salvar o pipeline")
    } finally {
      setSaving(false)
    }
  }

  return <Dialog open={open} onOpenChange={(next) => { if (!saving) onOpenChange(next) }}>
    <DialogContent className="min-w-0 overflow-x-hidden max-sm:p-4">
      <form noValidate onSubmit={submit} className="grid min-w-0 gap-5">
        <DialogHeader><DialogTitle>{pipeline ? "Editar pipeline" : "Novo pipeline"}</DialogTitle><DialogDescription>Integração entre uma origem e um destino de dados.</DialogDescription></DialogHeader>
        <nav aria-label="Etapas do pipeline" className="min-w-0"><p aria-live="polite" className="mb-3 text-sm font-semibold">Etapa {step + 1} de 6: {wizardSteps[step]}</p><ol className="grid min-w-0 grid-cols-2 gap-2 sm:grid-cols-3">{wizardSteps.map((label, index) => <li key={label} aria-current={step === index ? "step" : undefined} className={`flex min-w-0 items-start gap-2 rounded-md border p-2 text-xs ${step === index ? "border-primary bg-primary/5 font-semibold" : "border-border text-muted-foreground"}`}><span className="grid size-5 shrink-0 place-items-center rounded-full border border-current">{index + 1}</span><span className="min-w-0 break-words">{label}</span></li>)}</ol></nav>
        <fieldset disabled={saving} className="grid min-w-0 gap-4">
          <legend className="sr-only">{wizardSteps[step]}</legend>
          {step === 0 && <>
            <label className="grid min-w-0 gap-1.5 text-sm font-medium">Destino<select required value={destinationId} onChange={(event) => setDestinationId(event.target.value)} disabled={!destinations.length} className={inputClass}><option value="" disabled>Selecione o destino</option>{destinations.map((item) => <option key={item.id} value={item.id}>{item.name} - {technologyNames[item.kind] || item.kind}</option>)}</select></label>
            {!destinations.length && <p className="text-sm text-amber-900">Nenhum destino cadastrado. <a href="/destinations" className="font-semibold underline">Cadastrar destino</a></p>}
            {destination && <ConnectionSummary connection={destination} destination />}
          </>}
          {step === 1 && <>
            <label className="grid min-w-0 gap-1.5 text-sm font-medium">Conector de origem<select required value={sourceId} onChange={(event) => changeSource(event.target.value)} disabled={!sources.length} className={inputClass}><option value="" disabled>Selecione a origem</option>{sources.map((item) => <option key={item.id} value={item.id} disabled={!supportsDiscovery(item)}>{item.name} - {technologyNames[item.kind] || item.kind}{supportsDiscovery(item) ? "" : " (indisponível)"}</option>)}</select></label>
            {!sources.some(supportsDiscovery) && <p className="text-sm text-amber-900">Nenhuma origem PostgreSQL ou ClickHouse disponível. <a href="/connectors" className="font-semibold underline">Cadastrar conector</a></p>}
            {source && !discoverable && <p role="alert" className="text-sm text-amber-900">A descoberta de tabelas não está disponível para esta origem. Selecione PostgreSQL ou ClickHouse.</p>}
            {source && <ConnectionSummary connection={source} />}
          </>}
          {step === 2 && <>
            {!discoverable ? <p role="alert" className="text-sm text-amber-900">Selecione uma origem disponível na etapa anterior.</p> : tables?.key !== sourceId || tables.status === "loading" ? <p role="status" className="flex items-center gap-2 text-sm"><LoaderCircle className="size-4 animate-spin" />Carregando tabelas...</p> : tables.status === "error" ? <div role="alert" className="grid min-w-0 gap-3"><p className="break-words text-sm text-red-800">{tables.error}</p><Button type="button" variant="outline" onClick={() => setTablesRetry((current) => current + 1)}><RefreshCw />Tentar novamente</Button></div> : !availableTables.length ? <div className="grid gap-3"><p role="status" className="text-sm text-muted-foreground">Nenhuma tabela encontrada nesta origem.</p><Button type="button" variant="outline" onClick={() => setTablesRetry((current) => current + 1)}><RefreshCw />Atualizar tabelas</Button></div> : <>
              <label className="grid min-w-0 gap-1.5 text-sm font-medium">Tabela da origem<select required value={selectedTableKey} onChange={(event) => changeTable(event.target.value)} className={inputClass}><option value="" disabled>Selecione uma tabela</option>{sourceTable && !tableValid && <option value={selectedTableKey} disabled>{sourceSchema ? `${sourceSchema}.` : ""}{sourceTable} (indisponível)</option>}{availableTables.map((table) => <option key={tableKey(table)} value={tableKey(table)}>{table.schema ? `${table.schema}.` : ""}{table.name}</option>)}</select></label>
              {sourceTable && !tableValid && <p role="alert" className="text-sm text-amber-900">A tabela salva não foi encontrada. Selecione uma tabela disponível.</p>}
            </>}
          </>}
          {step === 3 && <label className="grid min-w-0 gap-1.5 text-sm font-medium">Nome da tabela nova no destino<input required maxLength={63} pattern="[A-Za-z_][A-Za-z0-9_]*" value={destinationTable} onChange={(event) => { setDestinationTable(event.target.value); setDestinationDerived(false) }} aria-invalid={!!destinationTable && !destinationTableValid} aria-describedby="destination-table-format" className={inputClass} placeholder="Ex.: vendas_warehouse" /><span id="destination-table-format" className="text-xs font-normal text-muted-foreground">Até 63 caracteres: letras ASCII, números e sublinhado. Comece com letra ou sublinhado.</span></label>}
          {step === 4 && <>
            {!discoverable || !tableValid ? <p role="alert" className="text-sm text-amber-900">Selecione uma tabela disponível nas etapas anteriores.</p> : columnMetadata?.key !== columnsKey || columnMetadata.status === "loading" ? <p role="status" className="flex items-center gap-2 text-sm"><LoaderCircle className="size-4 animate-spin" />Carregando colunas...</p> : columnMetadata.status === "error" ? <div role="alert" className="grid min-w-0 gap-3"><p className="break-words text-sm text-red-800">{columnMetadata.error}</p><Button type="button" variant="outline" onClick={() => setColumnsRetry((current) => current + 1)}><RefreshCw />Tentar novamente</Button></div> : !availableColumns.length ? <div className="grid gap-3"><p role="status" className="text-sm text-muted-foreground">Nenhuma coluna encontrada nesta tabela.</p><Button type="button" variant="outline" onClick={() => setColumnsRetry((current) => current + 1)}><RefreshCw />Atualizar colunas</Button></div> : <>
              <div className="flex min-w-0 flex-wrap items-center justify-between gap-2"><label className="flex min-w-0 items-center gap-2 text-sm font-medium"><input type="checkbox" checked={allColumnsSelected} ref={(element) => { if (element) element.indeterminate = !allColumnsSelected && availableColumns.some((column) => columns.includes(column.name)) }} onChange={(event) => setColumns(event.target.checked ? availableColumns.map((column) => column.name) : [])} className="size-4 shrink-0 accent-primary" />Todas as colunas</label><span role="status" className="text-xs text-muted-foreground">{columns.length} selecionada(s) de {availableColumns.length}</span></div>
              <div className="max-h-64 min-w-0 overflow-y-auto rounded-md border border-border">{availableColumns.map((column) => <label key={column.name} className="flex min-w-0 items-start gap-3 border-b border-border p-3 last:border-0"><input type="checkbox" checked={columns.includes(column.name)} onChange={(event) => setColumns((current) => event.target.checked ? [...current, column.name] : current.filter((name) => name !== column.name))} className="mt-0.5 size-4 shrink-0 accent-primary" /><span className="min-w-0"><span className="block break-all text-sm font-medium">{column.name}</span><span className="block break-all text-xs text-muted-foreground">{column.type}{column.nullable ? " · Aceita nulos" : " · Não nulo"}</span></span></label>)}</div>
              {unavailableColumns.length > 0 && <div className="grid gap-2 text-sm text-amber-900"><p role="alert">Colunas salvas indisponíveis. Desmarque-as para continuar.</p>{unavailableColumns.map((column) => <label key={column} className="flex min-w-0 items-start gap-2"><input type="checkbox" checked onChange={() => setColumns((current) => current.filter((name) => name !== column))} className="mt-0.5 size-4 shrink-0 accent-primary" /><span className="min-w-0 break-all">{column}</span></label>)}</div>}
              {!columns.length && <p className="text-sm text-muted-foreground">Selecione pelo menos uma coluna.</p>}
            </>}
          </>}
          {step === 5 && <>
            <label className="grid min-w-0 gap-1.5 text-sm font-medium">Nome do pipeline<input required minLength={2} maxLength={120} value={name} onChange={(event) => setName(event.target.value)} className={inputClass} placeholder="Ex.: Vendas para o warehouse" /></label>
            <label className="grid min-w-0 gap-1.5 text-sm font-medium">Frequência<select required value={frequency} onChange={(event) => setFrequency(event.target.value)} className={inputClass}>{!frequencies.some((item) => item.value === frequency) && <option value={frequency} disabled>Frequência indisponível</option>}{frequencies.map((item) => <option key={item.value} value={item.value}>{item.label}</option>)}</select></label>
            <dl className="grid min-w-0 gap-2 border-y border-border py-3 text-sm"><div className="min-w-0"><dt className="text-xs text-muted-foreground">Tabela da origem</dt><dd className="break-all">{sourceSchema ? `${sourceSchema}.` : ""}{sourceTable}</dd></div><div className="min-w-0"><dt className="text-xs text-muted-foreground">Tabela no destino</dt><dd className="break-all">{destinationTable}</dd></div><div><dt className="text-xs text-muted-foreground">Colunas selecionadas</dt><dd>{columns.length}</dd></div></dl>
          </>}
        </fieldset>
        {error && <div role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}</div>}
        <DialogFooter className="min-w-0 flex-wrap"><Button type="button" variant="ghost" disabled={saving} onClick={() => onOpenChange(false)}>Cancelar</Button>{step > 0 && <Button type="button" variant="outline" disabled={saving} onClick={() => { setStep((current) => current - 1); setError("") }}><ArrowLeft />Voltar</Button>}<Button type="submit" disabled={saving || (step === 5 ? !allValid : !validity[step])}>{saving ? <LoaderCircle className="animate-spin" /> : step === 5 ? <Check /> : <ArrowRight />}{saving ? "Salvando..." : step < 5 ? "Continuar" : pipeline ? "Salvar alterações" : "Salvar pipeline"}</Button></DialogFooter>
      </form>
    </DialogContent>
  </Dialog>
}

export function PipelinesPage() {
  const [pipelines, setPipelines] = useState<Pipeline[]>([])
  const [sources, setSources] = useState<Connection[]>([])
  const [destinations, setDestinations] = useState<Connection[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [query, setQuery] = useState("")
  const deferredQuery = useDeferredValue(query.trim().toLowerCase())
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingPipeline, setEditingPipeline] = useState<Pipeline | null>(null)
  const [pipelineToDelete, setPipelineToDelete] = useState<Pipeline | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState("")

  async function load() {
    setLoading(true)
    setError("")
    try {
      const [pipelineData, sourceData, destinationData] = await Promise.all([request<Pipeline[]>("/api/pipelines"), request<Connection[]>("/api/connectors"), request<Connection[]>("/api/destinations")])
      setPipelines(pipelineData)
      setSources(sourceData)
      setDestinations(destinationData)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Não foi possível carregar os pipelines")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function save(input: PipelineInput) {
    const data = await request<Pipeline>(editingPipeline ? `/api/pipelines/${editingPipeline.id}` : "/api/pipelines", { method: editingPipeline ? "PUT" : "POST", body: JSON.stringify(input) })
    setPipelines((current) => editingPipeline ? current.map((item) => item.id === data.id ? data : item) : [data, ...current])
  }

  async function remove() {
    if (!pipelineToDelete) return
    setDeleting(true)
    setDeleteError("")
    try {
      await request<void>(`/api/pipelines/${pipelineToDelete.id}`, { method: "DELETE" })
      setPipelines((current) => current.filter((item) => item.id !== pipelineToDelete.id))
      setPipelineToDelete(null)
    } catch (requestError) {
      setDeleteError(requestError instanceof Error ? requestError.message : "Não foi possível excluir o pipeline")
    } finally {
      setDeleting(false)
    }
  }

  const visiblePipelines = pipelines.filter((item) => `${item.name} ${item.description} ${item.source.name} ${item.destination.name} ${technologyNames[item.source.kind]} ${technologyNames[item.destination.kind]} ${item.sourceSchema || ""} ${item.sourceTable || ""} ${item.destinationTable || ""} ${frequencies.find((frequency) => frequency.value === item.frequency)?.label || ""}`.toLowerCase().includes(deferredQuery))

  return <>
    <section className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
      <div><p className="mb-1 text-sm font-medium text-primary">Integrações</p><h1 className="text-2xl font-bold sm:text-[28px]">Pipelines</h1><p className="mt-1 text-sm text-muted-foreground">Conecte suas origens aos destinos de dados.</p></div>
      <Button disabled={loading || !!error} onClick={() => { setEditingPipeline(null); setDialogOpen(true) }}><Plus />Novo pipeline</Button>
    </section>
    <section className="mb-4 grid gap-4 sm:grid-cols-3">
      {[{ label: "Pipelines", count: pipelines.length, icon: Workflow, colors: "bg-emerald-100 text-emerald-700" }, { label: "Origens em uso", count: new Set(pipelines.map((item) => item.source.id)).size, icon: Plug, colors: "bg-amber-100 text-amber-700" }, { label: "Destinos em uso", count: new Set(pipelines.map((item) => item.destination.id)).size, icon: DatabaseZap, colors: "bg-[#e8f0fe] text-[#2458a6]" }].map((metric) => <Card key={metric.label}><CardContent className="flex items-center gap-4 p-4"><span className={`grid size-10 place-items-center rounded-md ${metric.colors}`}><metric.icon className="size-5" /></span><div><strong className="text-xl">{metric.count}</strong><p className="text-xs text-muted-foreground">{metric.label}</p></div></CardContent></Card>)}
    </section>
    <Card>
      <div className="flex gap-3 border-b border-border p-4"><div className="relative flex-1"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><input value={query} onChange={(event) => setQuery(event.target.value)} className="h-9 w-full rounded-md border border-border bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" placeholder="Buscar pipelines..." /></div><Button variant="outline" size="icon" aria-label="Atualizar pipelines" title="Atualizar pipelines" disabled={loading} onClick={() => void load()}><RefreshCw className={loading ? "animate-spin" : ""} /></Button></div>
      <CardContent className="p-0">
        {loading ? <div className="grid min-h-56 place-items-center"><LoaderCircle className="size-6 animate-spin text-primary" /></div> : error ? <div className="grid min-h-56 place-items-center p-8 text-center"><div><TriangleAlert className="mx-auto mb-3 size-8 text-red-700" /><strong>Não foi possível carregar</strong><p className="mt-1 text-sm text-muted-foreground">{error}</p><Button variant="outline" className="mt-4" onClick={() => void load()}><RefreshCw />Tentar novamente</Button></div></div> : visiblePipelines.length ? <div className="divide-y divide-border">
          {visiblePipelines.map((pipeline) => <div key={pipeline.id} className="grid gap-4 p-4 hover:bg-muted/20 lg:grid-cols-[minmax(180px,0.7fr)_minmax(0,1fr)_auto] lg:items-center">
            <div className="min-w-0"><div className="flex flex-wrap items-center gap-2"><strong className="truncate text-sm" title={pipeline.name}>{pipeline.name}</strong><Badge variant="muted">Configurado</Badge></div>{pipeline.description && <p className="mt-1 break-words text-xs text-muted-foreground">{pipeline.description}</p>}</div>
            <div className="grid min-w-0 gap-2"><div className="grid min-w-0 items-center gap-3 sm:grid-cols-[minmax(0,1fr)_24px_minmax(0,1fr)]"><div className="min-w-0"><ConnectionSummary connection={pipeline.source} />{pipeline.sourceTable && <p className="mt-1 break-all text-xs text-muted-foreground">Origem: {pipeline.sourceSchema ? `${pipeline.sourceSchema}.` : ""}{pipeline.sourceTable}</p>}</div><ArrowRight className="size-5 text-muted-foreground max-sm:rotate-90" /><div className="min-w-0"><ConnectionSummary connection={pipeline.destination} destination />{pipeline.destinationTable && <p className="mt-1 break-all text-xs text-muted-foreground">Destino: {pipeline.destinationTable}</p>}</div></div>{pipeline.frequency && <p className="text-xs text-muted-foreground">Frequência: {frequencies.find((item) => item.value === pipeline.frequency)?.label || pipeline.frequency}</p>}</div>
            <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={`Ações de ${pipeline.name}`}><MoreHorizontal /></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem onSelect={() => { setEditingPipeline(pipeline); setDialogOpen(true) }}>Editar</DropdownMenuItem><DropdownMenuItem className="text-red-700 focus:text-red-800" onSelect={() => { setDeleteError(""); setPipelineToDelete(pipeline) }}>Excluir</DropdownMenuItem></DropdownMenuContent></DropdownMenu>
          </div>)}
        </div> : <div className="grid min-h-56 place-items-center p-8 text-center"><div><Workflow className="mx-auto mb-3 size-8 text-muted-foreground" /><strong>Nenhum pipeline encontrado</strong><p className="mt-1 text-sm text-muted-foreground">{query ? "Ajuste sua busca." : "Nenhuma integração cadastrada."}</p></div></div>}
      </CardContent>
    </Card>
    <PipelineDialog open={dialogOpen} pipeline={editingPipeline} sources={sources} destinations={destinations} onOpenChange={(open) => { setDialogOpen(open); if (!open) setEditingPipeline(null) }} onSave={save} />
    <Dialog open={pipelineToDelete !== null} onOpenChange={(open) => { if (!open && !deleting) setPipelineToDelete(null) }}>
      <DialogContent className="max-w-md"><DialogHeader><DialogTitle>Excluir pipeline?</DialogTitle><DialogDescription>O pipeline “{pipelineToDelete?.name}” será removido. O conector e o destino serão preservados.</DialogDescription></DialogHeader>{deleteError && <div role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{deleteError}</div>}<DialogFooter><Button variant="ghost" disabled={deleting} onClick={() => setPipelineToDelete(null)}>Cancelar</Button><Button className="bg-red-700 text-white hover:bg-red-800" disabled={deleting} onClick={() => void remove()}>{deleting ? <LoaderCircle className="animate-spin" /> : null}{deleting ? "Excluindo..." : "Excluir pipeline"}</Button></DialogFooter></DialogContent>
    </Dialog>
  </>
}