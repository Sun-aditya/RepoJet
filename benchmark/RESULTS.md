# RepoJet Benchmark Results (Step 7b)

Evaluated on 40 public GitHub repositories: **30 Tuning Set** and **10 Held-out Set**.

## Benchmark Metrics Comparison

| Metric | Tuning Set (30) | Held-out Set (10) | Overall (40) |
| :--- | :--- | :--- | :--- |
| **Clone Success** | 30 / 30 (100.0%) | 10 / 10 (100.0%) | 40 / 40 (100.0%) |
| **Analysis Success (of Cloned)** | 26 / 26 (100.0%) | 8 / 8 (100.0%) | 34 / 34 (100.0%) |
| **Framework Accuracy (of Analyzed)** | 25 / 26 (96.2%) | 8 / 8 (100.0%) | 33 / 34 (97.1%) |
| &nbsp;&nbsp;↳ *Applications* | 16 / 17 (94.1%) | 2 / 2 (100.0%) | 18 / 19 (94.7%) |
| &nbsp;&nbsp;↳ *Libraries* | 7 / 7 (100.0%) | 2 / 2 (100.0%) | 9 / 9 (100.0%) |
| &nbsp;&nbsp;↳ *Monorepos* | 2 / 2 (100.0%) | 4 / 4 (100.0%) | 6 / 6 (100.0%) |
| **Package Manager Accuracy (of Analyzed)** | 26 / 26 (100.0%) | 8 / 8 (100.0%) | 34 / 34 (100.0%) |
| **Project Type Accuracy (of Analyzed)** | 22 / 26 (84.6%) | 6 / 8 (75.0%) | 28 / 34 (82.4%) |
| **Clean Failure (Non-Node / No-Manifest)** | 4 / 4 (100.0%) | 2 / 2 (100.0%) | 6 / 6 (100.0%) |

## Repository Details

| Set | Category | Repository | Expected Fw | Detected Fw | Expected PM | Detected PM | Type | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| tuning | app | `vercel/next-learn` | Next.js | Next.js | pnpm | pnpm | application | PASS |
| tuning | app | `leerob/site` | Next.js | Next.js (+React) | pnpm | pnpm | application | PASS |
| tuning | app | `shadcn-ui/ui` | Next.js | Next.js (+Vite) | pnpm | pnpm | application | PASS |
| tuning | app | `steven-tey/precedent` | Next.js | Next.js (+React) | pnpm | pnpm | application | PASS |
| tuning | app | `vercel/platforms` | Next.js | Next.js (+React) | pnpm | pnpm | application | PASS |
| tuning | app | `vitejs/vite-plugin-react` | Vite | Vite | pnpm | pnpm | application | PASS |
| tuning | app | `antfu/vitesse-lite` | Vite | Vite | pnpm | pnpm | application | PASS |
| tuning | app | `sapphi-red/vite-plugin-static-copy` | Vite | Vite | pnpm | pnpm | application | PASS |
| tuning | app | `vikejs/vike` | Vite | Vite | pnpm | pnpm | application | PASS |
| tuning | app | `ElMassimo/vite_ruby` | Vite | Vite | pnpm | pnpm | unknown | PASS |
| tuning | app | `facebook/create-react-app` | none | None | npm | npm | application | PASS |
| tuning | app | `react-boilerplate/react-boilerplate` | React | React (+Express) | npm | npm | application | PASS |
| tuning | app | `davezuko/react-redux-starter-kit` | React | React (+Express) | yarn | yarn | application | PASS |
| tuning | app | `akveo/ngx-admin` | none | None | npm | npm | application | PASS |
| tuning | app | `expressjs/cors` | Express | None | npm | npm | library | MISMATCH |
| tuning | app | `hagopj13/node-express-boilerplate` | Express | Express | yarn | yarn | application | PASS |
| tuning | app | `santiq/bulletproof-nodejs` | Express | Express | npm | npm | application | PASS |
| tuning | library | `expressjs/express` | none | None | npm | npm | unknown | PASS |
| tuning | library | `expressjs/body-parser` | none | None | npm | npm | library | PASS |
| tuning | library | `chalk/chalk` | none | None | npm | npm | library | PASS |
| tuning | library | `sindresorhus/is` | none | None | npm | npm | library | PASS |
| tuning | library | `tj/commander.js` | none | None | npm | npm | library | PASS |
| tuning | library | `lodash/lodash` | none | None | npm | npm | library | PASS |
| tuning | library | `isaacs/node-glob` | none | None | npm | npm | library | PASS |
| tuning | monorepo | `vercel/turbo` | Next.js | Next.js (+Vite) | pnpm | pnpm | unknown | PASS |
| tuning | monorepo | `calcom/cal.com` | Next.js | Next.js (+Vite) | yarn | yarn | application | PASS |
| tuning | non-node | `pallets/flask` | - | - | - | - | - | PASS (clean failure) |
| tuning | non-node | `psf/requests` | - | - | - | - | - | PASS (clean failure) |
| tuning | non-node | `gin-gonic/gin` | - | - | - | - | - | PASS (clean failure) |
| tuning | non-node | `spf13/cobra` | - | - | - | - | - | PASS (clean failure) |
| held-out | app | `ixartz/Next-js-Boilerplate` | Next.js | Next.js (+React) | npm | npm | application | PASS |
| held-out | app | `antfu/vitesse` | Vite | Vite | pnpm | pnpm | application | PASS |
| held-out | library | `expressjs/morgan` | none | None | npm | npm | unknown | PASS |
| held-out | library | `axios/axios` | none | None | bun | bun | application | PASS |
| held-out | no-manifest | `enaqx/awesome-react` | - | - | - | - | - | PASS (clean failure) |
| held-out | monorepo | `trpc/trpc` | Next.js | Next.js (+Vite,React) | pnpm | pnpm | application | PASS |
| held-out | monorepo | `freeCodeCamp/freeCodeCamp` | none | None | pnpm | pnpm | application | PASS |
| held-out | monorepo | `colinhacks/zod` | Next.js | Next.js (+Vite) | npm | npm | application | PASS |
| held-out | monorepo | `storybookjs/storybook` | Next.js | Next.js (+Vite) | yarn | yarn | application | PASS |
| held-out | non-node | `BurntSushi/ripgrep` | - | - | - | - | - | PASS (clean failure) |

## Failures & Discrepancies

- **https://github.com/expressjs/cors** (tuning): Expected framework "Express" / PM "npm", detected framework "None" / PM "npm"
