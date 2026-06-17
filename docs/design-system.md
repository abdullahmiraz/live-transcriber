# Design System — Live Meet

> Source of truth for UI tokens, typography, motion, and layout patterns.
> **Read this before changing frontend visuals.** Update when tokens or patterns change.

## 1. Design direction

Live Meet is a **practical meeting product** — calm, readable, and consistent across pages.
Visual language: modern SaaS (Linear/Meet-inspired), not flashy.

| Principle | Rule |
|---|---|
| Clarity | High contrast body text; muted text only for secondary info |
| Consistency | Use CSS variables + shared layout components — no one-off hex colors |
| Motion | Subtle enter animations only; no heavy parallax or long transitions |
| Theme | **Light + dark** via `mode-watcher` (default: system preference) |
| Video UI | In-call room uses a dedicated dark **video surface** regardless of theme |

## 2. Color palette

All colors are defined as OKLCH CSS variables in `frontend/src/app.css`.
**Do not hardcode colors in components** — use semantic Tailwind classes.

### Semantic tokens

| Token | Usage |
|---|---|
| `background` / `foreground` | Page canvas and primary text |
| `card` / `card-foreground` | Elevated panels, sidebar in call |
| `primary` / `primary-foreground` | CTAs, brand accent, own chat bubbles |
| `secondary` | Secondary buttons, badges |
| `muted` / `muted-foreground` | Subtle fills, helper text, others' chat bubbles |
| `accent` | Hover states, highlights |
| `destructive` | Errors, leave/end, muted mic/cam |
| `success` | Translated captions, positive states |
| `border` / `input` / `ring` | Borders, inputs, focus rings |
| `surface-elevated` | Homepage/lobby cards (`.surface-card`) |
| `video-surface` | In-call background and camera preview wells |

### Brand gradient

`.text-gradient-brand` — primary → success (indigo → teal). Use sparingly on hero headlines only.

### Light mode (default `:root`)

- Background: soft slate `oklch(0.985 0.004 247)`
- Primary: indigo `oklch(0.52 0.19 264)`
- Success/teal: `oklch(0.62 0.14 165)`
- Muted foreground: `oklch(0.48 0.035 257)` — **minimum readable gray**

### Dark mode (`.dark`)

- Background: deep slate `oklch(0.16 0.025 265)` — not pure black
- Primary: lighter indigo `oklch(0.68 0.16 264)`
- Muted foreground: `oklch(0.65 0.03 257)`

## 3. Typography

| Role | Font | Tailwind |
|---|---|---|
| UI / headings | **Plus Jakarta Sans** | `font-sans` (default on `body`) |
| Codes / slugs | **JetBrains Mono** | `font-mono` |

Loaded in `frontend/src/app.html`. Headings use `tracking-tight` and slightly tighter letter-spacing via base CSS.

### Scale (common)

| Element | Classes |
|---|---|
| Hero H1 | `text-4xl md:text-5xl font-bold tracking-tight leading-[1.1]` |
| Page title | `text-xl font-semibold` |
| Body | `text-sm` or `text-base leading-relaxed` |
| Helper | `text-muted-foreground text-xs` |
| Meeting slug | `font-mono text-[0.65rem]` |

## 4. Spacing & layout

- Page max width: `max-w-6xl` (marketing), `max-w-lg` (lobby/error)
- Page padding: `px-5`
- Card padding: shadcn Card defaults + `surface-card` utility
- In-call sidebar width: `380px` at `lg` breakpoint
- Video tiles: `rounded-2xl`, `ring-1 ring-white/10` on video surface

### Shared layout components

| Component | Path | Use |
|---|---|---|
| `BrandLogo` | `lib/components/layout/BrandLogo.svelte` | Logo + wordmark; props: `compact`, `inverted` |
| `ThemeToggle` | `lib/components/layout/ThemeToggle.svelte` | Light/dark toggle |
| `AppHeader` | `lib/components/layout/AppHeader.svelte` | Logo + theme + optional slot |
| `PageStage` | `lib/components/layout/PageStage.svelte` | Route-keyed page enter animation |

## 5. Motion

Defined in `app.css` — lightweight CSS keyframes only (no animation library).

| Utility | Duration | Use |
|---|---|---|
| `.animate-page-enter` | 400ms fade-in-up | Route changes (`PageStage`) |
| `.animate-fade-in-up` | 450ms | Hero sections |
| `.animate-scale-in` | 300ms | Cards/modals appearing |
| `.animate-stagger` | 40ms steps | Feature lists (children nth-child delays) |
| `.animate-pulse-soft` | 2.2s loop | Live interim caption indicator |

**Do not** add bounce, spin (except loaders), or transitions longer than 500ms for page UI.

## 6. Component patterns

### shadcn-svelte first

Use vendored components in `src/lib/components/ui/`. Custom markup only for `<video>` tiles and ambient backgrounds.

### Cards

Marketing/lobby: `surface-card` class (elevated shadow + border).

### In-call control bar

`.meeting-control-bar` — frosted pill bar at bottom with rounded-full buttons.

### Chat bubbles

- Own messages: `bg-primary text-primary-foreground`, aligned right
- Others: `bg-muted`, aligned left
- Timestamps: `text-muted-foreground text-[0.65rem]`

### Captions

- Original: `text-sm leading-relaxed`
- Translation: `text-success italic`
- Container: `rounded-xl border bg-muted/40 p-3`

## 7. Theme switching

- `ModeWatcher` in `routes/+layout.svelte`, `defaultMode="system"`
- Storage key: `mode-watcher-mode` (localStorage)
- Toaster follows mode via `mode-watcher` in `sonner.svelte`
- **Do not** hardcode `class="dark"` on `<html>` in `app.html`

## 8. Page-specific notes

| Page | Shell |
|---|---|
| `/` (home) | `AppHeader` + ambient background |
| `/m/[slug]` lobby/error | `AppHeader` + `surface-card` |
| `/m/[slug]` in-call | Full-bleed `bg-video-surface`; sidebar uses theme `card` colors |

## 9. Checklist before shipping UI changes

- [ ] Colors use semantic tokens only
- [ ] Typography matches scale above
- [ ] New shared UI extracted to `lib/components/layout/` if reused
- [ ] Light **and** dark modes checked
- [ ] Page enter animation preserved on navigation
- [ ] `docs/design-system.md` updated if tokens/patterns change
- [ ] `docs/project-memory.md` ADR if decision is architectural

## 10. References

- Tokens: `frontend/src/app.css`
- Agent: `agents/design-system.md`
- Frontend agent: `agents/frontend-engineer.md`
