import { FormEvent, useDeferredValue, useEffect, useState } from "react"
import { Check, Database, LoaderCircle, MoreHorizontal, Plus, RefreshCw, Search, TriangleAlert } from "lucide-react"

import { authTokenKey } from "@/auth-page"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent } from "@/components/ui/card"
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from "@/components/ui/dialog"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { cn } from "@/lib/utils"

type ConnectorKind = "bigquery" | "mysql" | "postgresql" | "sqlserver" | "s3" | "rest"
type ConnectorStatus = "configured" | "attention"

type ConnectorDefinition = {
  kind: ConnectorKind
  name: string
  description: string
  mark: string
  colors: string
  defaultPort?: string
}

type Connector = {
  id: number
  name: string
  kind: ConnectorKind
  detail: string
  status: ConnectorStatus
  lastSync: string
}

type ConnectorInput = {
  name: string
  kind: ConnectorKind
  host: string
  port: number | null
  database: string
}

const definitions: ConnectorDefinition[] = [
  { kind: "bigquery", name: "Google BigQuery", description: "Data warehouse do Google Cloud", mark: "BQ", colors: "bg-[#e8f0fe] text-[#2458a6]" },
  { kind: "mysql", name: "MySQL", description: "Banco de dados relacional", mark: "my", colors: "bg-[#e7f4f7] text-[#176477]", defaultPort: "3306" },
  { kind: "postgresql", name: "PostgreSQL", description: "Banco de dados objeto-relacional", mark: "PG", colors: "bg-[#e8eef5] text-[#315d82]", defaultPort: "5432" },
  { kind: "sqlserver", name: "SQL Server", description: "Banco de dados Microsoft", mark: "SQL", colors: "bg-[#f7e9e7] text-[#9b3d35]", defaultPort: "1433" },
  { kind: "s3", name: "Amazon S3", description: "Armazenamento de objetos", mark: "S3", colors: "bg-[#fff1dc] text-[#9a5b09]" },
  { kind: "rest", name: "REST API", description: "Fonte HTTP personalizada", mark: "{ }", colors: "bg-[#e9eee9] text-[#315c47]" },
]

function ConnectorMark({ definition, size = "default" }: { definition: ConnectorDefinition; size?: "default" | "large" }) {
  return <span className={cn("grid shrink-0 place-items-center rounded-md font-bold", definition.colors, size === "large" ? "size-12 text-xs" : "size-10 text-[11px]")}>{definition.mark}</span>
}

function ConnectorDialog({ open, onOpenChange, onCreate }: { open: boolean; onOpenChange: (open: boolean) => void; onCreate: (input: ConnectorInput) => Promise<void> }) {
  const [selected, setSelected] = useState<ConnectorDefinition | null>(null)
  const [name, setName] = useState("")
  const [host, setHost] = useState("")
  const [port, setPort] = useState("")
  const [database, setDatabase] = useState("")
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  function reset() {
    setSelected(null)
    setName("")
    setHost("")
    setPort("")
    setDatabase("")
    setError("")
  }

  function choose(definition: ConnectorDefinition) {
    setSelected(definition)
    setName(definition.name)
    setPort(definition.defaultPort ?? "")
  }

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!selected || !name.trim()) return
    setError("")
    setSaving(true)
    try {
      await onCreate({ name: name.trim(), kind: selected.kind, host: host.trim(), port: port ? Number(port) : null, database: database.trim() })
      onOpenChange(false)
      reset()
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Não foi possível salvar o conector")
    } finally {
      setSaving(false)
    }
  }

  return (
    <Dialog open={open} onOpenChange={(next) => { onOpenChange(next); if (!next) reset() }}>
      <DialogContent>
        {!selected ? (
          <>
            <DialogHeader><DialogTitle>Novo conector</DialogTitle><DialogDescription>Escolha a tecnologia da origem de dados.</DialogDescription></DialogHeader>
            <div className="grid gap-3 sm:grid-cols-2">
              {definitions.map((definition) => (
                <button key={definition.kind} type="button" onClick={() => choose(definition)} className="flex min-h-20 items-center gap-3 rounded-md border border-border bg-card p-3 text-left transition-colors hover:border-primary/50 hover:bg-muted/40 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring">
                  <ConnectorMark definition={definition} /><span><strong className="block text-sm">{definition.name}</strong><span className="mt-0.5 block text-xs text-muted-foreground">{definition.description}</span></span>
                </button>
              ))}
            </div>
          </>
        ) : (
          <form onSubmit={submit} className="grid gap-5">
            <DialogHeader><div className="flex items-center gap-3"><ConnectorMark definition={selected} /><div><DialogTitle>Conectar {selected.name}</DialogTitle><DialogDescription>Informe os dados de acesso à origem.</DialogDescription></div></div></DialogHeader>
            <div className="grid gap-4">
              <label className="grid gap-1.5 text-sm font-medium">Nome do conector<input required value={name} onChange={(event) => setName(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="Ex.: Produção" /></label>
              {selected.kind === "bigquery" ? (
                <div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Project ID<input required value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="meu-projeto" /></label><label className="grid gap-1.5 text-sm font-medium">Dataset<input required value={database} onChange={(event) => setDatabase(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="analytics" /></label></div>
              ) : selected.kind === "rest" ? (
                <label className="grid gap-1.5 text-sm font-medium">URL base<input required type="url" value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="https://api.exemplo.com/v1" /></label>
              ) : (
                <><div className="grid gap-4 sm:grid-cols-[1fr_120px]"><label className="grid gap-1.5 text-sm font-medium">Host<input required value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="db.internal" /></label><label className="grid gap-1.5 text-sm font-medium">Porta<input value={port} onChange={(event) => setPort(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" /></label></div><label className="grid gap-1.5 text-sm font-medium">Banco de dados<input required value={database} onChange={(event) => setDatabase(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="database" /></label></>
              )}
              <div className="rounded-md border border-border bg-muted/40 p-3 text-xs text-muted-foreground">O conector será salvo como configurado. O teste real da origem será adicionado na próxima etapa.</div>
              {error && <div role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}</div>}
            </div>
            <DialogFooter><Button type="button" variant="outline" onClick={() => setSelected(null)} disabled={saving}>Voltar</Button><DialogClose asChild><Button type="button" variant="ghost" disabled={saving}>Cancelar</Button></DialogClose><Button type="submit" disabled={saving}>{saving ? <LoaderCircle className="animate-spin" /> : <Check />}{saving ? "Salvando..." : "Salvar conector"}</Button></DialogFooter>
          </form>
        )}
      </DialogContent>
    </Dialog>
  )
}

export function ConnectorsPage() {
  const [connectors, setConnectors] = useState<Connector[]>([])
  const [query, setQuery] = useState("")
  const [filter, setFilter] = useState<"all" | ConnectorStatus>("all")
  const [dialogOpen, setDialogOpen] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const deferredQuery = useDeferredValue(query.toLowerCase())

  async function request(path: string, init?: RequestInit) {
    const token = localStorage.getItem(authTokenKey)
    const response = await fetch(path, { ...init, headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, ...init?.headers } })
    const payload = await response.json() as { data?: Connector | Connector[]; message?: string }
    if (response.status === 401) {
      localStorage.removeItem(authTokenKey)
      window.location.reload()
      throw new Error("Sua sessão expirou")
    }
    if (!response.ok) throw new Error(payload.message || "Não foi possível acessar os conectores")
    return payload.data
  }

  async function loadConnectors() {
    setLoading(true)
    setError("")
    try {
      const data = await request("/api/connectors")
      setConnectors(data as Connector[])
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Não foi possível carregar os conectores")
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void loadConnectors()
  }, [])

  async function createConnector(input: ConnectorInput) {
    const data = await request("/api/connectors", { method: "POST", body: JSON.stringify(input) })
    setConnectors((current) => [data as Connector, ...current])
  }

  const visibleConnectors = connectors.filter((connector) => {
    const definition = definitions.find((item) => item.kind === connector.kind)!
    return (filter === "all" || connector.status === filter) && `${connector.name} ${definition.name} ${connector.detail}`.toLowerCase().includes(deferredQuery)
  })

  return (
    <>
      <section className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div><p className="mb-1 text-sm font-medium text-primary">Integrações</p><h1 className="text-2xl font-bold sm:text-[28px]">Conectores</h1><p className="mt-1 text-sm text-muted-foreground">Gerencie as origens que alimentam seus pipelines de dados.</p></div>
        <Button onClick={() => setDialogOpen(true)}><Plus />Novo conector</Button>
      </section>

      <section className="mb-4 grid gap-4 sm:grid-cols-3">
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-emerald-100 text-emerald-700"><Check className="size-5" /></span><div><strong className="text-xl">{connectors.filter((item) => item.status === "configured").length}</strong><p className="text-xs text-muted-foreground">Configurados</p></div></CardContent></Card>
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-amber-100 text-amber-700"><TriangleAlert className="size-5" /></span><div><strong className="text-xl">{connectors.filter((item) => item.status === "attention").length}</strong><p className="text-xs text-muted-foreground">Requer atenção</p></div></CardContent></Card>
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-[#e8f0fe] text-[#2458a6]"><RefreshCw className="size-5" /></span><div><strong className="text-xl">Nenhuma</strong><p className="text-xs text-muted-foreground">Última sincronização</p></div></CardContent></Card>
      </section>

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-center">
          <div className="relative flex-1"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><input value={query} onChange={(event) => setQuery(event.target.value)} className="h-9 w-full rounded-md border border-border bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" placeholder="Buscar conectores..." /></div>
          <div className="flex rounded-md border border-border p-1">{(["all", "configured", "attention"] as const).map((value) => <button key={value} type="button" onClick={() => setFilter(value)} className={cn("h-7 rounded px-3 text-xs font-semibold text-muted-foreground", filter === value && "bg-muted text-foreground")}>{value === "all" ? "Todos" : value === "configured" ? "Configurados" : "Atenção"}</button>)}</div>
        </div>
        <CardContent className="p-0">
          {loading ? <div className="grid min-h-56 place-items-center"><LoaderCircle className="size-6 animate-spin text-primary" /></div> : error ? <div className="grid min-h-56 place-items-center p-8 text-center"><div><TriangleAlert className="mx-auto mb-3 size-8 text-red-700" /><strong>Não foi possível carregar</strong><p className="mt-1 text-sm text-muted-foreground">{error}</p><Button className="mt-4" variant="outline" onClick={() => void loadConnectors()}><RefreshCw />Tentar novamente</Button></div></div> : visibleConnectors.length ? <div className="divide-y divide-border">{visibleConnectors.map((connector) => {
            const definition = definitions.find((item) => item.kind === connector.kind)!
            return <div key={connector.id} className="grid gap-4 p-4 hover:bg-muted/20 sm:grid-cols-[auto_minmax(0,1fr)_auto_auto] sm:items-center"><ConnectorMark definition={definition} size="large" /><div className="min-w-0"><div className="flex items-center gap-2"><strong className="truncate text-sm">{connector.name}</strong><Badge variant={connector.status === "configured" ? "success" : "warning"}>{connector.status === "configured" ? "Configurado" : "Atenção"}</Badge></div><p className="mt-1 truncate text-xs text-muted-foreground">{definition.name} · {connector.detail}</p></div><div className="text-left sm:text-right"><p className="text-xs text-muted-foreground">Última sincronização</p><p className="mt-0.5 text-sm font-medium">{connector.lastSync}</p></div><DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={`Ações de ${connector.name}`}><MoreHorizontal /></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem>Testar conexão</DropdownMenuItem><DropdownMenuItem>Editar</DropdownMenuItem><DropdownMenuItem>Ver execuções</DropdownMenuItem></DropdownMenuContent></DropdownMenu></div>
          })}</div> : <div className="grid min-h-56 place-items-center p-8 text-center"><div><Database className="mx-auto mb-3 size-8 text-muted-foreground" /><strong>Nenhum conector encontrado</strong><p className="mt-1 text-sm text-muted-foreground">Ajuste sua busca ou adicione uma nova origem.</p></div></div>}
        </CardContent>
      </Card>
      <ConnectorDialog open={dialogOpen} onOpenChange={setDialogOpen} onCreate={createConnector} />
    </>
  )
}