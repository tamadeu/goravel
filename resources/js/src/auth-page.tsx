import { FormEvent, useState } from "react"
import { ArrowRight, Check, Command, Database, Eye, EyeOff, GitBranch, LoaderCircle, LockKeyhole } from "lucide-react"

import { Button } from "@/components/ui/button"

export type AuthUser = {
  ID: number
  name: string
  email: string
}

type AuthPayload = {
  token: string
  token_type: string
  user: AuthUser
}

export const authTokenKey = "goravel_data_token"

export function AuthPage({ onAuthenticated }: { onAuthenticated: (user: AuthUser) => void }) {
  const [mode, setMode] = useState<"login" | "register">("login")
  const [name, setName] = useState("")
  const [email, setEmail] = useState("")
  const [password, setPassword] = useState("")
  const [passwordConfirmation, setPasswordConfirmation] = useState("")
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState("")
  const [submitting, setSubmitting] = useState(false)

  async function submit(event: FormEvent) {
    event.preventDefault()
    setError("")
    setSubmitting(true)

    try {
      const response = await fetch(`/api/auth/${mode === "login" ? "login" : "register"}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(mode === "login" ? { email, password } : { name, email, password, password_confirmation: passwordConfirmation }),
      })
      const payload = await response.json() as AuthPayload & { message?: string }
      if (!response.ok) throw new Error(payload.message || "Não foi possível continuar")

      localStorage.setItem(authTokenKey, payload.token)
      onAuthenticated(payload.user)
    } catch (requestError) {
      setError(requestError instanceof Error ? requestError.message : "Não foi possível continuar")
    } finally {
      setSubmitting(false)
    }
  }

  function changeMode(nextMode: "login" | "register") {
    setMode(nextMode)
    setError("")
    setPassword("")
    setPasswordConfirmation("")
  }

  return (
    <main className="grid min-h-screen bg-[#f7f7f5] lg:grid-cols-[minmax(360px,0.8fr)_minmax(520px,1.2fr)]">
      <section className="relative hidden overflow-hidden border-r border-[#264e42] bg-[#174f40] p-10 text-white lg:flex lg:flex-col">
        <div className="absolute inset-0 opacity-15 [background-image:radial-gradient(#fff_0.7px,transparent_0.7px)] [background-size:18px_18px]" />
        <a href="/dashboard/" className="relative flex items-center gap-2.5 text-lg font-bold"><span className="grid size-9 place-items-center rounded-md bg-white text-[#174f40]"><Command className="size-5" /></span>Goravel Data</a>
        <div className="relative my-auto max-w-md">
          <p className="mb-4 text-xs font-bold uppercase text-emerald-200">Data integration workspace</p>
          <h1 className="text-4xl font-bold leading-tight">Seus dados, conectados em um só fluxo.</h1>
          <p className="mt-5 max-w-sm text-base leading-relaxed text-emerald-50/75">Conecte origens, transforme informações e acompanhe cada execução com clareza.</p>
          <div className="mt-10 grid grid-cols-[44px_1fr_44px_1fr_44px] items-center gap-3" aria-hidden="true">
            <span className="grid size-11 place-items-center rounded-md border border-white/20 bg-white/10"><Database className="size-5" /></span><span className="h-px bg-emerald-200/40" /><span className="grid size-11 place-items-center rounded-md border border-white/20 bg-white/10"><GitBranch className="size-5" /></span><span className="h-px bg-emerald-200/40" /><span className="grid size-11 place-items-center rounded-md bg-white text-[#174f40]"><Check className="size-5" /></span>
          </div>
        </div>
        <p className="relative text-xs text-emerald-100/60">Integração segura para operações de dados.</p>
      </section>

      <section className="flex min-h-screen items-center justify-center p-5 sm:p-10">
        <div className="w-full max-w-[420px]">
          <div className="mb-9 flex items-center gap-2.5 font-bold lg:hidden"><span className="grid size-8 place-items-center rounded-md bg-primary text-primary-foreground"><Command className="size-4" /></span>Goravel Data</div>
          <div className="mb-7">
            <p className="mb-2 text-sm font-semibold text-primary">{mode === "login" ? "Bem-vindo de volta" : "Crie seu workspace"}</p>
            <h2 className="text-3xl font-bold">{mode === "login" ? "Entre na sua conta" : "Comece agora"}</h2>
            <p className="mt-2 text-sm text-muted-foreground">{mode === "login" ? "Use suas credenciais para acessar o dashboard." : "Cadastre-se para configurar sua primeira integração."}</p>
          </div>

          <form onSubmit={submit} className="grid gap-4">
            {mode === "register" && <label className="grid gap-1.5 text-sm font-semibold">Nome<input autoComplete="name" required value={name} onChange={(event) => setName(event.target.value)} className="h-11 rounded-md border border-border bg-white px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="Seu nome" /></label>}
            <label className="grid gap-1.5 text-sm font-semibold">Email<input type="email" autoComplete="email" required value={email} onChange={(event) => setEmail(event.target.value)} className="h-11 rounded-md border border-border bg-white px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="voce@empresa.com" /></label>
            <label className="grid gap-1.5 text-sm font-semibold">Senha<span className="relative"><input type={showPassword ? "text" : "password"} autoComplete={mode === "login" ? "current-password" : "new-password"} minLength={8} required value={password} onChange={(event) => setPassword(event.target.value)} className="h-11 w-full rounded-md border border-border bg-white px-3 pr-11 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="Mínimo de 8 caracteres" /><button type="button" onClick={() => setShowPassword((visible) => !visible)} className="absolute right-1 top-1 grid size-9 place-items-center rounded text-muted-foreground hover:bg-muted" aria-label={showPassword ? "Ocultar senha" : "Mostrar senha"}>{showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}</button></span></label>
            {mode === "register" && <label className="grid gap-1.5 text-sm font-semibold">Confirme a senha<input type="password" autoComplete="new-password" minLength={8} required value={passwordConfirmation} onChange={(event) => setPasswordConfirmation(event.target.value)} className="h-11 rounded-md border border-border bg-white px-3 font-normal outline-none focus:ring-2 focus:ring-ring" placeholder="Repita sua senha" /></label>}

            {error && <div role="alert" className="flex items-start gap-2 rounded-md border border-red-200 bg-red-50 p-3 text-sm text-red-800"><LockKeyhole className="mt-0.5 size-4 shrink-0" />{error}</div>}
            <Button type="submit" className="mt-1 h-11" disabled={submitting}>{submitting ? <><LoaderCircle className="animate-spin" />Aguarde...</> : <>{mode === "login" ? "Entrar" : "Criar conta"}<ArrowRight /></>}</Button>
          </form>

          <p className="mt-6 text-center text-sm text-muted-foreground">{mode === "login" ? "Ainda não tem uma conta?" : "Já possui uma conta?"} <button type="button" className="font-semibold text-primary hover:underline" onClick={() => changeMode(mode === "login" ? "register" : "login")}>{mode === "login" ? "Cadastre-se" : "Entrar"}</button></p>
        </div>
      </section>
    </main>
  )
}