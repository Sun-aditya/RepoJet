# RepoJet Benchmark Results

Run on 40 public GitHub repositories.

## Summary

| Metric | Result |
| :--- | :--- |
| **Total Repositories** | 40 |
| **Node.js Repositories** | 35 |
| **Successful Node.js Analyses** | 31 / 35 (88.6%) |
| **Framework Detection Accuracy** | 19 / 35 (54.3%) |
| **Package Manager Accuracy** | 28 / 35 (80.0%) |
| **Non-Node Repositories** | 5 |
| **Non-Node Clean Failure Rate** | 5 / 5 (100.0%) |
| **Average Analysis Duration** | 0 ms |

## Repository Details

| Repository | Expected Framework | Detected Framework | Expected PM | Detected PM | Duration | Status |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| `vercel/next-learn` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `leerob/site` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `shadcn-ui/ui` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `steven-tey/precedent` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `mheap/nextjs-starter-template` | Next.js | - | npm | - | 0ms | ERROR (clone_error) |
| `vercel/platforms` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `vitejs/vite-plugin-react` | Vite | Vite | pnpm | pnpm | 0ms | PASS |
| `antfu/vitesse-lite` | Vite | Vite | pnpm | pnpm | 0ms | PASS |
| `sapphi-red/vite-plugin-static-copy` | Vite | Vite | pnpm | pnpm | 0ms | PASS |
| `sitepoint/vite-starter-template` | Vite | - | npm | - | 0ms | ERROR (clone_error) |
| `alex-streza/vite-react-ts-tailwind-starter` | Vite | - | npm | - | 0ms | ERROR (clone_error) |
| `ElMassimo/vite_ruby` | Vite | Vite | pnpm | pnpm | 0ms | PASS |
| `facebook/create-react-app` | React | None | yarn | npm | 0ms | MISMATCH |
| `react-boilerplate/react-boilerplate` | React | Express | npm | npm | 0ms | MISMATCH |
| `davezuko/react-redux-starter-kit` | React | Express | yarn | yarn | 0ms | MISMATCH |
| `enaqx/awesome-react` | None | - | npm | - | 0ms | ERROR (unsupported_project_type) |
| `akveo/ngx-admin` | React | None | npm | npm | 0ms | MISMATCH |
| `expressjs/express` | Express | None | npm | npm | 0ms | MISMATCH |
| `expressjs/cors` | Express | Express | npm | npm | 0ms | PASS |
| `expressjs/body-parser` | Express | None | npm | npm | 0ms | MISMATCH |
| `expressjs/morgan` | Express | None | npm | npm | 0ms | MISMATCH |
| `hagopj13/node-express-boilerplate` | Express | Express | yarn | yarn | 0ms | PASS |
| `santiq/bulletproof-nodejs` | Express | Express | npm | npm | 0ms | PASS |
| `chalk/chalk` | None | None | npm | npm | 0ms | PASS |
| `sindresorhus/is` | None | None | npm | npm | 0ms | PASS |
| `tj/commander.js` | None | None | npm | npm | 0ms | PASS |
| `lodash/lodash` | None | None | npm | npm | 0ms | PASS |
| `axios/axios` | None | Express | npm | bun | 0ms | MISMATCH |
| `isaacs/node-glob` | None | None | npm | npm | 0ms | PASS |
| `vercel/turbo` | Next.js | Next.js | pnpm | pnpm | 0ms | PASS |
| `trpc/trpc` | None | Next.js | pnpm | pnpm | 0ms | MISMATCH |
| `calcom/cal.com` | Next.js | Next.js | yarn | yarn | 0ms | PASS |
| `freeCodeCamp/freeCodeCamp` | Next.js | None | pnpm | pnpm | 0ms | MISMATCH |
| `colinhacks/zod` | None | Next.js | yarn | npm | 0ms | MISMATCH |
| `storybookjs/storybook` | React | Next.js | yarn | yarn | 0ms | MISMATCH |
| `pallets/flask` | Non-Node | - | Non-Node | - | 0ms | PASS (clean failure) |
| `psf/requests` | Non-Node | - | Non-Node | - | 0ms | PASS (clean failure) |
| `gin-gonic/gin` | Non-Node | - | Non-Node | - | 0ms | PASS (clean failure) |
| `spf13/cobra` | Non-Node | - | Non-Node | - | 0ms | PASS (clean failure) |
| `BurntSushi/ripgrep` | Non-Node | - | Non-Node | - | 0ms | PASS (clean failure) |

## Failures & Discrepancies

- **https://github.com/mheap/nextjs-starter-template**: Failed with clone_error: failed to clone repository: exit status 128: Cloning into '/tmp/repojet-3084907813'...
fatal: could not read Username for 'https://github.com': terminal prompts disabled

- **https://github.com/sitepoint/vite-starter-template**: Failed with clone_error: failed to clone repository: exit status 128: Cloning into '/tmp/repojet-3243114628'...
fatal: could not read Username for 'https://github.com': terminal prompts disabled

- **https://github.com/alex-streza/vite-react-ts-tailwind-starter**: Failed with clone_error: failed to clone repository: exit status 128: Cloning into '/tmp/repojet-3193692520'...
fatal: could not read Username for 'https://github.com': terminal prompts disabled

- **https://github.com/facebook/create-react-app**: Expected React / yarn, detected None / npm
- **https://github.com/react-boilerplate/react-boilerplate**: Expected React / npm, detected Express / npm
- **https://github.com/davezuko/react-redux-starter-kit**: Expected React / yarn, detected Express / yarn
- **https://github.com/enaqx/awesome-react**: Failed with unsupported_project_type: unsupported project type: no matching analyzer found
- **https://github.com/akveo/ngx-admin**: Expected React / npm, detected None / npm
- **https://github.com/expressjs/express**: Expected Express / npm, detected None / npm
- **https://github.com/expressjs/body-parser**: Expected Express / npm, detected None / npm
- **https://github.com/expressjs/morgan**: Expected Express / npm, detected None / npm
- **https://github.com/axios/axios**: Expected None / npm, detected Express / bun
- **https://github.com/trpc/trpc**: Expected None / pnpm, detected Next.js / pnpm
- **https://github.com/freeCodeCamp/freeCodeCamp**: Expected Next.js / pnpm, detected None / pnpm
- **https://github.com/colinhacks/zod**: Expected None / yarn, detected Next.js / npm
- **https://github.com/storybookjs/storybook**: Expected React / yarn, detected Next.js / yarn
