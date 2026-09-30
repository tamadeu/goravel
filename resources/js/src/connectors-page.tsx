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
type ConnectorStatus = "configured" | "connected" | "attention"

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
  configuration: {
    host: string
    port: number | null
    database: string
    username: string
    schema: string
    sslMode: string
    connectTimeout: number
    hasPassword: boolean
    hasSSLRootCert: boolean
    hasSSLClientCertificate: boolean
  }
}

type ConnectorInput = {
  name: string
  kind: ConnectorKind
  host: string
  port: number | null
  database: string
  username: string
  password: string
  schema: string
  sslMode: string
  sslRootCert: string
  sslCertificate: string
  sslKey: string
  connectTimeout: number
  clearPassword: boolean
  clearSSLRootCert: boolean
  clearSSLClientCertificate: boolean
}

type PostgreSQLTestResult = {
  database: string
  user: string
  version: string
  latencyMs: number
}

type ConnectorTestResponse = {
  result: PostgreSQLTestResult
  connector: Connector
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

function ConnectorDialog({ open, editingConnector, onOpenChange, onSave }: { open: boolean; editingConnector: Connector | null; onOpenChange: (open: boolean) => void; onSave: (input: ConnectorInput) => Promise<void> }) {
  const [selected, setSelected] = useState<ConnectorDefinition | null>(null)
  const [name, setName] = useState("")
  const [host, setHost] = useState("")
  const [port, setPort] = useState("")
  const [database, setDatabase] = useState("")
  const [username, setUsername] = useState("")
  const [password, setPassword] = useState("")
  const [schema, setSchema] = useState("public")
  const [sslMode, setSSLMode] = useState("prefer")
  const [sslRootCert, setSSLRootCert] = useState("")
  const [sslCertificate, setSSLCertificate] = useState("")
  const [sslKey, setSSLKey] = useState("")
  const [connectTimeout, setConnectTimeout] = useState("10")
  const [clearPassword, setClearPassword] = useState(false)
  const [clearSSLRootCert, setClearSSLRootCert] = useState(false)
  const [clearSSLClientCertificate, setClearSSLClientCertificate] = useState(false)
  const [error, setError] = useState("")
  const [saving, setSaving] = useState(false)

  function reset() {
    setSelected(null)
    setName("")
    setHost("")
    setPort("")
    setDatabase("")
    setUsername("")
    setPassword("")
    setSchema("public")
    setSSLMode("prefer")
    setSSLRootCert("")
    setSSLCertificate("")
    setSSLKey("")
    setConnectTimeout("10")
    setClearPassword(false)
    setClearSSLRootCert(false)
    setClearSSLClientCertificate(false)
    setError("")
  }

  function choose(definition: ConnectorDefinition) {
    setSelected(definition)
    setName(definition.name)
    setPort(definition.defaultPort ?? "")
  }

  useEffect(() => {
    if (!open || !editingConnector) return
    const definition = definitions.find((item) => item.kind === editingConnector.kind) ?? null
    setSelected(definition)
    setName(editingConnector.name)
    setHost(editingConnector.configuration.host)
    setPort(editingConnector.configuration.port?.toString() ?? "")
    setDatabase(editingConnector.configuration.database)
    setUsername(editingConnector.configuration.username)
    setPassword("")
    setSchema(editingConnector.configuration.schema || "public")
    setSSLMode(editingConnector.configuration.sslMode || "prefer")
    setSSLRootCert("")
    setSSLCertificate("")
    setSSLKey("")
    setConnectTimeout((editingConnector.configuration.connectTimeout || 10).toString())
    setClearPassword(false)
    setClearSSLRootCert(false)
    setClearSSLClientCertificate(false)
    setError("")
  }, [editingConnector, open])

  async function submit(event: FormEvent) {
    event.preventDefault()
    if (!selected || !name.trim()) return
    setError("")
    setSaving(true)
    try {
      await onSave({
        name: name.trim(), kind: selected.kind, host: host.trim(), port: port ? Number(port) : null, database: database.trim(),
        username: username.trim(), password, schema: schema.trim(), sslMode, sslRootCert: sslRootCert.trim(),
        sslCertificate: sslCertificate.trim(), sslKey: sslKey.trim(), connectTimeout: Number(connectTimeout),
        clearPassword, clearSSLRootCert, clearSSLClientCertificate,
      })
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
            <DialogHeader><div className="flex items-center gap-3"><ConnectorMark definition={selected} /><div><DialogTitle>{editingConnector ? `Editar ${editingConnector.name}` : `Conectar ${selected.name}`}</DialogTitle><DialogDescription>{editingConnector ? "Atualize a configuração da origem de dados." : "Informe os dados de acesso à origem."}</DialogDescription></div></div></DialogHeader>
            <div className="grid gap-4">
              <label className="grid gap-1.5 text-sm font-medium">Nome do conector<input required value={name} onChange={(event) => setName(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="Ex.: Produção" /></label>
              {selected.kind === "postgresql" ? (
                <>
                  <div className="grid gap-4 sm:grid-cols-[1fr_120px]"><label className="grid gap-1.5 text-sm font-medium">Host<input required value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="postgres.exemplo.com" /></label><label className="grid gap-1.5 text-sm font-medium">Porta<input required type="number" min="1" max="65535" value={port} onChange={(event) => setPort(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" /></label></div>
                  <div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Banco de dados<input required value={database} onChange={(event) => setDatabase(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="database" /></label><label className="grid gap-1.5 text-sm font-medium">Schema<input required value={schema} onChange={(event) => setSchema(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="public" /></label></div>
                  <div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Usuário<input required autoComplete="username" value={username} onChange={(event) => setUsername(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="postgres" /></label><label className="grid gap-1.5 text-sm font-medium">Senha<input type="password" autoComplete="new-password" value={password} onChange={(event) => { setPassword(event.target.value); if (event.target.value) setClearPassword(false) }} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder={editingConnector?.configuration.hasPassword ? "Deixe vazio para manter a atual" : "Opcional para autenticação trust"} />{editingConnector?.configuration.hasPassword && <span className="flex items-center gap-2 text-xs font-normal text-muted-foreground"><input type="checkbox" checked={clearPassword} onChange={(event) => { setClearPassword(event.target.checked); if (event.target.checked) setPassword("") }} />Remover senha armazenada</span>}</label></div>
                  <details className="rounded-md border border-border bg-muted/20 p-4"><summary className="cursor-pointer text-sm font-semibold">SSL e opções avançadas</summary><div className="mt-4 grid gap-4"><div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Modo SSL<select value={sslMode} onChange={(event) => setSSLMode(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring"><option value="disable">disable</option><option value="allow">allow</option><option value="prefer">prefer</option><option value="require">require</option><option value="verify-ca">verify-ca</option><option value="verify-full">verify-full</option></select></label><label className="grid gap-1.5 text-sm font-medium">Timeout de conexão<input type="number" min="1" max="60" value={connectTimeout} onChange={(event) => setConnectTimeout(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" /><span className="text-xs font-normal text-muted-foreground">Em segundos, entre 1 e 60.</span></label></div>{sslMode !== "disable" && <><label className="grid gap-1.5 text-sm font-medium">Certificado raiz (CA)<textarea value={sslRootCert} onChange={(event) => { setSSLRootCert(event.target.value); if (event.target.value) setClearSSLRootCert(false) }} className="min-h-20 resize-y rounded-md border border-border bg-background p-3 font-mono text-xs font-normal outline-none focus:ring-2 focus:ring-ring" placeholder={editingConnector?.configuration.hasSSLRootCert ? "Deixe vazio para manter o certificado atual" : "-----BEGIN CERTIFICATE-----"} />{editingConnector?.configuration.hasSSLRootCert && <span className="flex items-center gap-2 text-xs font-normal text-muted-foreground"><input type="checkbox" checked={clearSSLRootCert} onChange={(event) => { setClearSSLRootCert(event.target.checked); if (event.target.checked) setSSLRootCert("") }} />Remover certificado raiz</span>}</label><div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Certificado do cliente<textarea value={sslCertificate} onChange={(event) => { setSSLCertificate(event.target.value); if (event.target.value) setClearSSLClientCertificate(false) }} className="min-h-24 resize-y rounded-md border border-border bg-background p-3 font-mono text-xs font-normal outline-none focus:ring-2 focus:ring-ring" placeholder={editingConnector?.configuration.hasSSLClientCertificate ? "Deixe vazio para manter o atual" : "-----BEGIN CERTIFICATE-----"} />{editingConnector?.configuration.hasSSLClientCertificate && <span className="flex items-center gap-2 text-xs font-normal text-muted-foreground"><input type="checkbox" checked={clearSSLClientCertificate} onChange={(event) => { setClearSSLClientCertificate(event.target.checked); if (event.target.checked) { setSSLCertificate(""); setSSLKey("") } }} />Remover certificado e chave</span>}</label><label className="grid gap-1.5 text-sm font-medium">Chave do cliente<textarea value={sslKey} onChange={(event) => { setSSLKey(event.target.value); if (event.target.value) setClearSSLClientCertificate(false) }} className="min-h-24 resize-y rounded-md border border-border bg-background p-3 font-mono text-xs font-normal outline-none focus:ring-2 focus:ring-ring" placeholder={editingConnector?.configuration.hasSSLClientCertificate ? "Deixe vazio para manter a atual" : "-----BEGIN PRIVATE KEY-----"} /></label></div></>}</div></details>
                </>
              ) : selected.kind === "bigquery" ? (
                <div className="grid gap-4 sm:grid-cols-2"><label className="grid gap-1.5 text-sm font-medium">Project ID<input required value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="meu-projeto" /></label><label className="grid gap-1.5 text-sm font-medium">Dataset<input required value={database} onChange={(event) => setDatabase(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="analytics" /></label></div>
              ) : selected.kind === "rest" ? (
                <label className="grid gap-1.5 text-sm font-medium">URL base<input required type="url" value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="https://api.exemplo.com/v1" /></label>
              ) : (
                <><div className="grid gap-4 sm:grid-cols-[1fr_120px]"><label className="grid gap-1.5 text-sm font-medium">Host<input required value={host} onChange={(event) => setHost(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="db.internal" /></label><label className="grid gap-1.5 text-sm font-medium">Porta<input value={port} onChange={(event) => setPort(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" /></label></div><label className="grid gap-1.5 text-sm font-medium">Banco de dados<input required value={database} onChange={(event) => setDatabase(event.target.value)} className="h-10 rounded-md border border-border bg-background px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="database" /></label></>
              )}
              <div className="rounded-md border border-border bg-muted/40 p-3 text-xs text-muted-foreground">Senha, certificados e chaves serão armazenados criptografados. Após salvar, use a ação “Testar conexão”.</div>
              {error && <div role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{error}</div>}
            </div>
            <DialogFooter>{!editingConnector && <Button type="button" variant="outline" onClick={() => setSelected(null)} disabled={saving}>Voltar</Button>}<DialogClose asChild><Button type="button" variant="ghost" disabled={saving}>Cancelar</Button></DialogClose><Button type="submit" disabled={saving}>{saving ? <LoaderCircle className="animate-spin" /> : <Check />}{saving ? "Salvando..." : editingConnector ? "Salvar alterações" : "Salvar conector"}</Button></DialogFooter>
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
  const [editingConnector, setEditingConnector] = useState<Connector | null>(null)
  const [connectorToDelete, setConnectorToDelete] = useState<Connector | null>(null)
  const [deleting, setDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState("")
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState("")
  const [testState, setTestState] = useState<{ connector: Connector; status: "loading" | "success" | "error"; result?: PostgreSQLTestResult; message?: string } | null>(null)
  const deferredQuery = useDeferredValue(query.toLowerCase())

  async function request<T>(path: string, init?: RequestInit): Promise<T> {
    const token = localStorage.getItem(authTokenKey)
    const response = await fetch(path, { ...init, headers: { "Content-Type": "application/json", Authorization: `Bearer ${token}`, ...init?.headers } })
    if (response.status === 401) {
      localStorage.removeItem(authTokenKey)
      window.location.reload()
      throw new Error("Sua sessão expirou")
    }
    if (response.status === 204) return undefined as T
    const payload = await response.json() as { data?: T; message?: string }
    if (!response.ok) throw new Error(payload.message || "Não foi possível acessar os conectores")
    return payload.data as T
  }

  async function loadConnectors() {
    setLoading(true)
    setError("")
    try {
      const data = await request<Connector[]>("/api/connectors")
      setConnectors(data)
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
    const data = await request<Connector>("/api/connectors", { method: "POST", body: JSON.stringify(input) })
    setConnectors((current) => [data, ...current])
  }

  async function updateConnector(input: ConnectorInput) {
    if (!editingConnector) return
    const data = await request<Connector>(`/api/connectors/${editingConnector.id}`, { method: "PUT", body: JSON.stringify(input) })
    setConnectors((current) => current.map((connector) => connector.id === data.id ? data : connector))
  }

  async function deleteConnector() {
    if (!connectorToDelete) return
    setDeleting(true)
    setDeleteError("")
    try {
      await request<void>(`/api/connectors/${connectorToDelete.id}`, { method: "DELETE" })
      setConnectors((current) => current.filter((connector) => connector.id !== connectorToDelete.id))
      setConnectorToDelete(null)
    } catch (requestError) {
      setDeleteError(requestError instanceof Error ? requestError.message : "Não foi possível excluir o conector")
    } finally {
      setDeleting(false)
    }
  }

  async function testConnection(connector: Connector) {
    setTestState({ connector, status: "loading" })
    try {
      const data = await request<ConnectorTestResponse>(`/api/connectors/${connector.id}/test`, { method: "POST" })
      setConnectors((current) => current.map((item) => item.id === connector.id ? data.connector : item))
      setTestState({ connector: data.connector, status: "success", result: data.result })
    } catch (requestError) {
      setConnectors((current) => current.map((item) => item.id === connector.id ? { ...item, status: "attention" } : item))
      setTestState({ connector, status: "error", message: requestError instanceof Error ? requestError.message : "Não foi possível testar a conexão" })
    }
  }

  const visibleConnectors = connectors.filter((connector) => {
    const definition = definitions.find((item) => item.kind === connector.kind)!
    const matchesFilter = filter === "all" || connector.status === filter || (filter === "configured" && connector.status === "connected")
    return matchesFilter && `${connector.name} ${definition.name} ${connector.detail}`.toLowerCase().includes(deferredQuery)
  })
  const latestTest = connectors.find((connector) => connector.status === "connected")?.lastSync ?? "Nenhum"

  return (
    <>
      <section className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
        <div><p className="mb-1 text-sm font-medium text-primary">Integrações</p><h1 className="text-2xl font-bold sm:text-[28px]">Conectores</h1><p className="mt-1 text-sm text-muted-foreground">Gerencie as origens que alimentam seus pipelines de dados.</p></div>
        <Button onClick={() => { setEditingConnector(null); setDialogOpen(true) }}><Plus />Novo conector</Button>
      </section>

      <section className="mb-4 grid gap-4 sm:grid-cols-3">
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-emerald-100 text-emerald-700"><Check className="size-5" /></span><div><strong className="text-xl">{connectors.filter((item) => item.status === "configured" || item.status === "connected").length}</strong><p className="text-xs text-muted-foreground">Configurados</p></div></CardContent></Card>
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-amber-100 text-amber-700"><TriangleAlert className="size-5" /></span><div><strong className="text-xl">{connectors.filter((item) => item.status === "attention").length}</strong><p className="text-xs text-muted-foreground">Requer atenção</p></div></CardContent></Card>
        <Card><CardContent className="flex items-center gap-4 p-4"><span className="grid size-10 place-items-center rounded-md bg-[#e8f0fe] text-[#2458a6]"><RefreshCw className="size-5" /></span><div><strong className="text-xl">{latestTest}</strong><p className="text-xs text-muted-foreground">Último teste</p></div></CardContent></Card>
      </section>

      <Card>
        <div className="flex flex-col gap-3 border-b border-border p-4 sm:flex-row sm:items-center">
          <div className="relative flex-1"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><input value={query} onChange={(event) => setQuery(event.target.value)} className="h-9 w-full rounded-md border border-border bg-background pl-9 pr-3 text-sm outline-none focus:ring-2 focus:ring-ring" placeholder="Buscar conectores..." /></div>
          <div className="flex rounded-md border border-border p-1">{(["all", "configured", "attention"] as const).map((value) => <button key={value} type="button" onClick={() => setFilter(value)} className={cn("h-7 rounded px-3 text-xs font-semibold text-muted-foreground", filter === value && "bg-muted text-foreground")}>{value === "all" ? "Todos" : value === "configured" ? "Configurados" : "Atenção"}</button>)}</div>
        </div>
        <CardContent className="p-0">
          {loading ? <div className="grid min-h-56 place-items-center"><LoaderCircle className="size-6 animate-spin text-primary" /></div> : error ? <div className="grid min-h-56 place-items-center p-8 text-center"><div><TriangleAlert className="mx-auto mb-3 size-8 text-red-700" /><strong>Não foi possível carregar</strong><p className="mt-1 text-sm text-muted-foreground">{error}</p><Button className="mt-4" variant="outline" onClick={() => void loadConnectors()}><RefreshCw />Tentar novamente</Button></div></div> : visibleConnectors.length ? <div className="divide-y divide-border">{visibleConnectors.map((connector) => {
            const definition = definitions.find((item) => item.kind === connector.kind)!
            const connected = connector.status === "connected"
            return <div key={connector.id} className="grid gap-4 p-4 hover:bg-muted/20 sm:grid-cols-[auto_minmax(0,1fr)_auto_auto] sm:items-center"><ConnectorMark definition={definition} size="large" /><div className="min-w-0"><div className="flex items-center gap-2"><strong className="truncate text-sm">{connector.name}</strong><Badge variant={connector.status === "attention" ? "warning" : "success"}>{connected ? "Conectado" : connector.status === "configured" ? "Configurado" : "Atenção"}</Badge></div><p className="mt-1 truncate text-xs text-muted-foreground">{definition.name} · {connector.detail}</p></div><div className="text-left sm:text-right"><p className="text-xs text-muted-foreground">Último teste</p><p className="mt-0.5 text-sm font-medium">{connector.lastSync}</p></div><DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" size="icon" aria-label={`Ações de ${connector.name}`}><MoreHorizontal /></Button></DropdownMenuTrigger><DropdownMenuContent align="end">{connector.kind === "postgresql" && <DropdownMenuItem onSelect={() => void testConnection(connector)}>Testar conexão</DropdownMenuItem>}<DropdownMenuItem onSelect={() => { setEditingConnector(connector); setDialogOpen(true) }}>Editar</DropdownMenuItem><DropdownMenuItem>Ver execuções</DropdownMenuItem><DropdownMenuItem className="text-red-700 focus:text-red-800" onSelect={() => { setDeleteError(""); setConnectorToDelete(connector) }}>Excluir</DropdownMenuItem></DropdownMenuContent></DropdownMenu></div>
          })}</div> : <div className="grid min-h-56 place-items-center p-8 text-center"><div><Database className="mx-auto mb-3 size-8 text-muted-foreground" /><strong>Nenhum conector encontrado</strong><p className="mt-1 text-sm text-muted-foreground">Ajuste sua busca ou adicione uma nova origem.</p></div></div>}
        </CardContent>
      </Card>
      <ConnectorDialog open={dialogOpen} editingConnector={editingConnector} onOpenChange={(open) => { setDialogOpen(open); if (!open) setEditingConnector(null) }} onSave={editingConnector ? updateConnector : createConnector} />
      <Dialog open={connectorToDelete !== null} onOpenChange={(open) => { if (!open && !deleting) setConnectorToDelete(null) }}>
        <DialogContent className="max-w-md">
          <DialogHeader><DialogTitle>Excluir conector?</DialogTitle><DialogDescription>O conector “{connectorToDelete?.name}” será removido permanentemente. Esta ação não pode ser desfeita.</DialogDescription></DialogHeader>
          {deleteError && <div role="alert" className="rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800">{deleteError}</div>}
          <DialogFooter><Button variant="ghost" onClick={() => setConnectorToDelete(null)} disabled={deleting}>Cancelar</Button><Button className="bg-red-700 text-white hover:bg-red-800" onClick={() => void deleteConnector()} disabled={deleting}>{deleting ? <LoaderCircle className="animate-spin" /> : null}{deleting ? "Excluindo..." : "Excluir conector"}</Button></DialogFooter>
        </DialogContent>
      </Dialog>
      <Dialog open={testState !== null} onOpenChange={(open) => { if (!open && testState?.status !== "loading") setTestState(null) }}>
        <DialogContent className="max-w-lg">
          <DialogHeader><DialogTitle>Teste de conexão</DialogTitle><DialogDescription>{testState?.connector.name}</DialogDescription></DialogHeader>
          {testState?.status === "loading" && <div className="grid min-h-40 place-items-center text-center"><div><LoaderCircle className="mx-auto size-7 animate-spin text-primary" /><p className="mt-3 text-sm text-muted-foreground">Conectando ao PostgreSQL...</p></div></div>}
          {testState?.status === "error" && <div className="rounded-md border border-red-200 bg-red-50 p-4"><div className="flex gap-3"><TriangleAlert className="mt-0.5 size-5 shrink-0 text-red-700" /><div><strong className="text-sm text-red-900">Não foi possível conectar</strong><p className="mt-1 break-words text-sm text-red-800">{testState.message}</p></div></div></div>}
          {testState?.status === "success" && testState.result && <div className="grid gap-3"><div className="flex items-center gap-3 rounded-md border border-emerald-200 bg-emerald-50 p-4"><span className="grid size-9 place-items-center rounded-full bg-emerald-100 text-emerald-700"><Check className="size-5" /></span><div><strong className="text-sm text-emerald-900">Conexão estabelecida</strong><p className="text-xs text-emerald-800">Resposta em {testState.result.latencyMs} ms</p></div></div><dl className="grid grid-cols-[120px_1fr] gap-x-4 gap-y-3 rounded-md border border-border p-4 text-sm"><dt className="text-muted-foreground">Database</dt><dd className="font-medium">{testState.result.database}</dd><dt className="text-muted-foreground">Usuário</dt><dd className="font-medium">{testState.result.user}</dd><dt className="text-muted-foreground">Versão</dt><dd className="break-words text-xs leading-relaxed">{testState.result.version}</dd></dl></div>}
          {testState?.status !== "loading" && <DialogFooter><Button onClick={() => setTestState(null)}>Fechar</Button></DialogFooter>}
        </DialogContent>
      </Dialog>
    </>
  )
}