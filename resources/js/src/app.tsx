import { useEffect, useState } from "react"
import {
  Activity,
  Bell,
  Box,
  ChevronDown,
  CircleHelp,
  Command,
  CreditCard,
  Download,
  LayoutDashboard,
  Menu,
  MoreHorizontal,
  Plug,
  Search,
  Settings,
  ShoppingBag,
  TrendingDown,
  TrendingUp,
  Users,
  Workflow,
  X,
} from "lucide-react"
import { Area, AreaChart, CartesianGrid, ResponsiveContainer, Tooltip as ChartTooltip, XAxis, YAxis } from "recharts"

import { Avatar, AvatarFallback } from "@/components/ui/avatar"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "@/components/ui/dropdown-menu"
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip"
import { AuthPage, type AuthUser, authTokenKey } from "@/auth-page"
import { ConnectorsPage } from "@/connectors-page"
import { cn } from "@/lib/utils"

const navigation = [
  { label: "Visão geral", icon: LayoutDashboard, page: "overview" as const, path: "/dashboard/" },
  { label: "Conectores", icon: Plug, page: "connectors" as const, path: "/dashboard/connectors" },
  { label: "Pipelines", icon: Workflow },
  { label: "Execuções", icon: Activity, count: 3 },
]

type Page = "overview" | "connectors"

function pageFromPath(): Page {
  return window.location.pathname === "/dashboard/connectors" ? "connectors" : "overview"
}

const chartData = [
  { day: "01 set", current: 18000, previous: 13400 },
  { day: "05 set", current: 22400, previous: 17600 },
  { day: "09 set", current: 19800, previous: 20200 },
  { day: "13 set", current: 28400, previous: 21900 },
  { day: "17 set", current: 31200, previous: 24700 },
  { day: "21 set", current: 27800, previous: 26000 },
  { day: "25 set", current: 36500, previous: 28100 },
  { day: "29 set", current: 42180, previous: 30400 },
]

const orders = [
  { id: "#1048", customer: "Marina Costa", initials: "MC", product: "Plano Pro anual", value: "R$ 1.188,00", status: "Pago", statusVariant: "success" as const, time: "há 4 min" },
  { id: "#1047", customer: "Felipe Nunes", initials: "FN", product: "Kit Workspace", value: "R$ 429,90", status: "Processando", statusVariant: "warning" as const, time: "há 18 min" },
  { id: "#1046", customer: "Ana Ribeiro", initials: "AR", product: "Plano Starter", value: "R$ 89,00", status: "Pago", statusVariant: "success" as const, time: "há 31 min" },
  { id: "#1045", customer: "Lucas Martins", initials: "LM", product: "Consultoria", value: "R$ 2.400,00", status: "Pendente", statusVariant: "muted" as const, time: "há 1 h" },
]

const channels = [
  { name: "Busca orgânica", value: 42, color: "bg-[#176b55]" },
  { name: "Indicações", value: 28, color: "bg-[#e0a328]" },
  { name: "Mídia paga", value: 18, color: "bg-[#3f6ac4]" },
  { name: "Direto", value: 12, color: "bg-[#c85a4a]" },
]

function formatCurrency(value: number) {
  return new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL", maximumFractionDigits: 0 }).format(value)
}

function Sidebar({ open, currentPage, user, onClose, onNavigate }: { open: boolean; currentPage: Page; user: AuthUser; onClose: () => void; onNavigate: (page: Page) => void }) {
  const initials = user.name.split(" ").map((part) => part[0]).join("").slice(0, 2).toUpperCase()
  return (
    <>
      <button className={cn("fixed inset-0 z-30 bg-black/30 transition-opacity lg:hidden", open ? "opacity-100" : "pointer-events-none opacity-0")} onClick={onClose} aria-label="Fechar menu" />
      <aside className="dashboard-sidebar fixed inset-y-0 left-0 z-40 flex w-64 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-transform" data-open={open}>
        <div className="flex h-16 items-center justify-between border-b border-sidebar-border px-5">
          <a href="/dashboard/" onClick={(event) => { event.preventDefault(); onNavigate("overview"); onClose() }} className="flex items-center gap-2.5 font-bold"><span className="grid size-8 place-items-center rounded-md bg-primary text-primary-foreground"><Command className="size-4" /></span><span>Goravel</span></a>
          <Button variant="ghost" size="icon" className="lg:hidden" onClick={onClose} aria-label="Fechar menu"><X /></Button>
        </div>
        <div className="flex-1 overflow-y-auto px-3 py-5">
          <p className="mb-2 px-3 text-[11px] font-bold uppercase text-muted-foreground">Workspace</p>
          <nav className="space-y-1">
            {navigation.map((item) => (
              <a key={item.label} href={item.path ?? "#"} onClick={(event) => { if (item.page) { event.preventDefault(); onNavigate(item.page) }; onClose() }} className={cn("flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium text-muted-foreground transition-colors hover:bg-sidebar-accent hover:text-sidebar-accent-foreground", item.page === currentPage && "bg-sidebar-accent text-sidebar-accent-foreground")}>
                <item.icon className="size-[18px]" /><span className="flex-1">{item.label}</span>{item.count && <span className="rounded-full bg-primary px-2 py-0.5 text-[10px] font-bold text-primary-foreground">{item.count}</span>}
              </a>
            ))}
          </nav>
          <p className="mb-2 mt-8 px-3 text-[11px] font-bold uppercase text-muted-foreground">Sistema</p>
          <nav className="space-y-1">
            <a href="#" className="flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"><Box className="size-[18px]" />Integrações</a>
            <a href="#" className="flex h-10 items-center gap-3 rounded-md px-3 text-sm font-medium text-muted-foreground hover:bg-sidebar-accent hover:text-sidebar-accent-foreground"><Settings className="size-[18px]" />Configurações</a>
          </nav>
        </div>
        <div className="border-t border-sidebar-border p-3">
          <div className="flex items-center gap-3 rounded-md p-2"><Avatar><AvatarFallback className="bg-[#e7d9c6] text-[#704c30]">{initials}</AvatarFallback></Avatar><div className="min-w-0 flex-1"><p className="truncate text-sm font-semibold">{user.name}</p><p className="truncate text-xs text-muted-foreground">{user.email}</p></div><MoreHorizontal className="size-4 text-muted-foreground" /></div>
        </div>
      </aside>
    </>
  )
}

function MetricCard({ title, value, change, down, icon: Icon }: { title: string; value: string; change: string; down?: boolean; icon: typeof Users }) {
  return (
    <Card className="animate-rise">
      <CardContent className="p-5">
        <div className="mb-5 flex items-start justify-between"><div className="grid size-9 place-items-center rounded-md border border-border bg-muted/50"><Icon className="size-[17px] text-muted-foreground" /></div><Badge variant={down ? "warning" : "success"} className="gap-1">{down ? <TrendingDown className="size-3" /> : <TrendingUp className="size-3" />}{change}</Badge></div>
        <p className="text-sm text-muted-foreground">{title}</p><p className="mt-1 text-2xl font-bold tracking-tight">{value}</p>
      </CardContent>
    </Card>
  )
}

export function App() {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const [currentPage, setCurrentPage] = useState<Page>(pageFromPath)
  const [authState, setAuthState] = useState<"loading" | "guest" | "authenticated">("loading")
  const [user, setUser] = useState<AuthUser | null>(null)

  useEffect(() => {
    const token = localStorage.getItem(authTokenKey)
    if (!token) {
      setAuthState("guest")
      return
    }

    fetch("/api/auth/me", { headers: { Authorization: `Bearer ${token}` } })
      .then(async (response) => {
        if (!response.ok) throw new Error("Invalid session")
        return response.json() as Promise<{ user: AuthUser }>
      })
      .then((payload) => {
        setUser(payload.user)
        setAuthState("authenticated")
      })
      .catch(() => {
        localStorage.removeItem(authTokenKey)
        setAuthState("guest")
      })
  }, [])

  useEffect(() => {
    const handlePopState = () => setCurrentPage(pageFromPath())
    window.addEventListener("popstate", handlePopState)
    return () => window.removeEventListener("popstate", handlePopState)
  }, [])

  function navigate(page: Page) {
    const path = page === "connectors" ? "/dashboard/connectors" : "/dashboard/"
    if (window.location.pathname !== path) window.history.pushState({}, "", path)
    setCurrentPage(page)
  }

  async function logout() {
    const token = localStorage.getItem(authTokenKey)
    if (token) await fetch("/api/auth/logout", { method: "POST", headers: { Authorization: `Bearer ${token}` } })
    localStorage.removeItem(authTokenKey)
    setUser(null)
    setAuthState("guest")
  }

  if (authState === "loading") {
    return <div className="grid min-h-screen place-items-center bg-background"><div className="size-7 animate-spin rounded-full border-2 border-muted border-t-primary" aria-label="Carregando sessão" /></div>
  }

  if (authState === "guest" || !user) {
    return <AuthPage onAuthenticated={(authenticatedUser) => { setUser(authenticatedUser); setAuthState("authenticated") }} />
  }

  const initials = user.name.split(" ").map((part) => part[0]).join("").slice(0, 2).toUpperCase()

  return (
    <TooltipProvider>
      <div className="min-h-screen bg-background text-foreground">
        <Sidebar open={sidebarOpen} currentPage={currentPage} user={user} onClose={() => setSidebarOpen(false)} onNavigate={navigate} />
        <div className="lg:pl-64">
          <header className="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-border bg-background/95 px-4 backdrop-blur md:px-7">
            <Button variant="ghost" size="icon" className="lg:hidden" onClick={() => setSidebarOpen(true)} aria-label="Abrir menu"><Menu /></Button>
            <div className="relative hidden w-full max-w-sm md:block"><Search className="absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" /><input className="h-9 w-full rounded-md border border-border bg-muted/40 pl-9 pr-3 text-sm outline-none placeholder:text-muted-foreground focus:ring-2 focus:ring-ring" placeholder="Buscar conectores, pipelines..." /></div>
            <div className="ml-auto flex items-center gap-1">
              <Tooltip><TooltipTrigger asChild><Button variant="ghost" size="icon" aria-label="Ajuda"><CircleHelp /></Button></TooltipTrigger><TooltipContent>Central de ajuda</TooltipContent></Tooltip>
              <Tooltip><TooltipTrigger asChild><Button variant="ghost" size="icon" className="relative" aria-label="Notificações"><Bell /><span className="absolute right-2 top-2 size-1.5 rounded-full bg-[#c85a4a]" /></Button></TooltipTrigger><TooltipContent>3 notificações</TooltipContent></Tooltip>
              <div className="mx-2 h-6 w-px bg-border" />
              <DropdownMenu><DropdownMenuTrigger asChild><Button variant="ghost" className="gap-2 px-2"><Avatar className="size-7"><AvatarFallback className="bg-[#e7d9c6] text-[#704c30]">{initials}</AvatarFallback></Avatar><span className="hidden sm:inline">{user.name.split(" ")[0]}</span><ChevronDown className="size-3.5 text-muted-foreground" /></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuLabel>{user.email}</DropdownMenuLabel><DropdownMenuSeparator /><DropdownMenuItem>Perfil</DropdownMenuItem><DropdownMenuItem>Preferências</DropdownMenuItem><DropdownMenuSeparator /><DropdownMenuItem onSelect={logout}>Sair</DropdownMenuItem></DropdownMenuContent></DropdownMenu>
            </div>
          </header>

          <main className="mx-auto max-w-[1500px] p-4 md:p-7">
            {currentPage === "connectors" ? <ConnectorsPage /> : <>
            <section className="mb-7 flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
              <div><p className="mb-1 text-sm font-medium text-primary">Terça-feira, 29 de setembro</p><h1 className="text-2xl font-bold sm:text-[28px]">Visão geral</h1><p className="mt-1 text-sm text-muted-foreground">Acompanhe o desempenho da sua operação.</p></div>
              <div className="flex items-center gap-2"><DropdownMenu><DropdownMenuTrigger asChild><Button variant="outline">Últimos 30 dias<ChevronDown /></Button></DropdownMenuTrigger><DropdownMenuContent align="end"><DropdownMenuItem>Últimos 7 dias</DropdownMenuItem><DropdownMenuItem>Últimos 30 dias</DropdownMenuItem><DropdownMenuItem>Este ano</DropdownMenuItem></DropdownMenuContent></DropdownMenu><Button><Download />Exportar</Button></div>
            </section>

            <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
              <MetricCard title="Receita total" value="R$ 84.240" change="12,5%" icon={CreditCard} />
              <MetricCard title="Novos pedidos" value="1.429" change="8,2%" icon={ShoppingBag} />
              <MetricCard title="Novos clientes" value="384" change="5,7%" icon={Users} />
              <MetricCard title="Ticket médio" value="R$ 312,40" change="2,1%" down icon={TrendingUp} />
            </section>

            <section className="mt-4 grid gap-4 xl:grid-cols-[minmax(0,1.7fr)_minmax(300px,0.8fr)]">
              <Card className="animate-rise [animation-delay:100ms]">
                <CardHeader className="flex-row items-start justify-between gap-4"><div><CardTitle>Receita</CardTitle><CardDescription className="mt-1">Comparativo mensal de vendas</CardDescription></div><div className="flex gap-4 text-xs text-muted-foreground"><span className="flex items-center gap-1.5"><i className="size-2 rounded-full bg-primary" />Atual</span><span className="flex items-center gap-1.5"><i className="size-2 rounded-full bg-[#c7c9c5]" />Anterior</span></div></CardHeader>
                <CardContent className="h-[300px] pl-1 pr-4 sm:h-[340px]">
                  <ResponsiveContainer width="100%" height="100%"><AreaChart data={chartData} margin={{ top: 10, right: 8, left: 0, bottom: 0 }}><defs><linearGradient id="revenueFill" x1="0" y1="0" x2="0" y2="1"><stop offset="0%" stopColor="#176b55" stopOpacity={0.22} /><stop offset="100%" stopColor="#176b55" stopOpacity={0} /></linearGradient></defs><CartesianGrid vertical={false} stroke="#e7e7e3" strokeDasharray="3 3" /><XAxis dataKey="day" tickLine={false} axisLine={false} tick={{ fill: "#767873", fontSize: 11 }} dy={10} /><YAxis tickLine={false} axisLine={false} width={58} tick={{ fill: "#767873", fontSize: 11 }} tickFormatter={(value) => `${value / 1000}k`} /><ChartTooltip formatter={(value) => formatCurrency(Number(value))} contentStyle={{ borderRadius: 6, borderColor: "#deded9", fontSize: 12 }} /><Area type="monotone" dataKey="previous" stroke="#b9bbb6" strokeWidth={2} fill="transparent" strokeDasharray="5 5" /><Area type="monotone" dataKey="current" stroke="#176b55" strokeWidth={2.5} fill="url(#revenueFill)" /></AreaChart></ResponsiveContainer>
                </CardContent>
              </Card>

              <Card className="animate-rise [animation-delay:150ms]">
                <CardHeader><CardTitle>Aquisição</CardTitle><CardDescription>Origem dos novos clientes</CardDescription></CardHeader>
                <CardContent>
                  <div className="mb-7 flex items-baseline gap-2"><span className="text-3xl font-bold">12.840</span><span className="text-sm text-muted-foreground">visitas</span></div>
                  <div className="mb-7 flex h-3 overflow-hidden rounded-full">{channels.map((channel) => <span key={channel.name} className={channel.color} style={{ width: `${channel.value}%` }} />)}</div>
                  <div className="space-y-4">{channels.map((channel) => <div key={channel.name} className="grid grid-cols-[12px_1fr_auto] items-center gap-2.5"><span className={cn("size-2.5 rounded-full", channel.color)} /><span className="text-sm text-muted-foreground">{channel.name}</span><strong className="text-sm">{channel.value}%</strong></div>)}</div>
                  <div className="mt-7 border-t border-border pt-5"><div className="flex items-center justify-between text-sm"><span className="text-muted-foreground">Taxa de conversão</span><strong className="text-emerald-700">3,24%</strong></div></div>
                </CardContent>
              </Card>
            </section>

            <Card className="mt-4 animate-rise [animation-delay:200ms]">
              <CardHeader className="flex-row items-center justify-between"><div><CardTitle>Pedidos recentes</CardTitle><CardDescription className="mt-1">Últimas transações da loja</CardDescription></div><Button variant="outline" size="sm">Ver todos</Button></CardHeader>
              <CardContent className="px-0 pb-0"><div className="overflow-x-auto"><table className="w-full min-w-[760px] text-left text-sm"><thead><tr className="border-y border-border bg-muted/40 text-xs text-muted-foreground"><th className="px-5 py-3 font-medium">Pedido</th><th className="px-5 py-3 font-medium">Cliente</th><th className="px-5 py-3 font-medium">Produto</th><th className="px-5 py-3 font-medium">Status</th><th className="px-5 py-3 text-right font-medium">Valor</th><th className="px-5 py-3 font-medium"><span className="sr-only">Ações</span></th></tr></thead><tbody>{orders.map((order) => <tr key={order.id} className="border-b border-border last:border-0 hover:bg-muted/30"><td className="px-5 py-4 font-semibold">{order.id}<span className="mt-0.5 block text-xs font-normal text-muted-foreground">{order.time}</span></td><td className="px-5 py-4"><div className="flex items-center gap-3"><Avatar className="size-8"><AvatarFallback>{order.initials}</AvatarFallback></Avatar><span className="font-medium">{order.customer}</span></div></td><td className="px-5 py-4 text-muted-foreground">{order.product}</td><td className="px-5 py-4"><Badge variant={order.statusVariant}>{order.status}</Badge></td><td className="px-5 py-4 text-right font-semibold">{order.value}</td><td className="px-5 py-4"><Button variant="ghost" size="icon" aria-label={`Ações do pedido ${order.id}`}><MoreHorizontal /></Button></td></tr>)}</tbody></table></div></CardContent>
            </Card>
            </>}
          </main>
        </div>
      </div>
    </TooltipProvider>
  )
}