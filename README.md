# Dark Patterns in Game Reviews (dsr)

This is the code I built for my master thesis. The idea is to find dark patterns (manipulative design
tricks) in video games by reading what players write in their public reviews, instead of looking at
screenshots of the game.

It works like a small pipeline. I collect reviews from Steam (and Reddit too), then a panel of language
models reads each review and says which dark patterns it sees. After that I check a sample of those
decisions by hand, so I have a human reference to measure the panel against.

## What it does

- Collects public reviews for a list of games.
- Has a panel of a few LLMs annotate each review against a taxonomy of 19 dark patterns (the MESO level).
- Lets me adjudicate a sample by hand: for each pattern I confirm or replace the panel's majority vote.
- Keeps everything frozen and versioned (the sample, the prompt, the taxonomy, the panel), so a run can be
  reproduced later.
- Lets me compare two runs to see where the models disagree, and look at which patterns each run flags the
  most.

## How it works

The data goes through a few steps. Each one leaves a trace in the database.

1. **Scrape** the reviews for a game from Steam or Reddit. The pages are pulled in the background.
2. **Curate** a population. This freezes a fixed sample of reviews (so many per game), so a run always
   annotates the same set even if new reviews arrive later.
3. **Run + annotate**. A run pins a population, a prompt, a taxonomy version and a panel of annotators.
   Then the panel reads every review and stores its answers.
4. **Adjudicate**. I open a sample in the frontend and decide each pattern by hand. There is an open pass
   (I see the panel votes) and a blind pass (I don't). The panel state is saved at the moment I decide, so
   the decision stays auditable.
5. **Compare**. I can put two runs side by side and look at the divergence per review, per model and per
   pattern.

The main objects are: games, reviews (artifacts), populations, runs, annotators (human or LLM), and the
adjudications.

## What's inside

**Backend** (Go 1.25). Postgres for storage, with `sqlc` to generate the queries and `pgx` to talk to the
database. Background jobs run on [River](https://riverqueue.com). The LLM calls go through
[OpenRouter](https://openrouter.ai), so I can use models from different families in the same panel.

**Frontend** (`frontend/`). Next.js 16 with React 19, Tailwind and TanStack Query. It is the console where
I manage games, populations and runs, and where I do the adjudication. It talks to the Go API over HTTP.

## Running it locally

You need: Go 1.25, Node, Docker, and [golang-migrate](https://github.com/golang-migrate/migrate) (the
`migrate` command). You only need `sqlc` if you change the SQL.

1. Start Postgres:

   ```bash
   docker-compose up -d
   ```

   This runs Postgres 16 on `localhost:5436` (database `dsr_db`, user `dsr_usr`).

2. Create your `.env` from the example and add your OpenRouter key:

   ```bash
   cp .env.example .env
   # then add this line to .env:
   # OPENROUTER_API_KEY=sk-or-...
   ```

3. Run the migrations and set up the job queue tables:

   ```bash
   make migrate-up
   make river-up
   ```

4. Start the API (port 8080):

   ```bash
   go run ./cmd/api
   ```

5. Start the workers, each in its own terminal:

   ```bash
   go run ./cmd/scrape serve
   go run ./cmd/annotate serve
   go run ./cmd/research serve
   ```

6. Start the frontend:

   ```bash
   cd frontend
   npm install
   npm run dev
   ```

   Open http://localhost:3000. The frontend reads the API address from `NEXT_PUBLIC_API_URL`
   (`frontend/.env.local`), which is `http://localhost:8080` by default.

If you just want to test the pipeline without spending money on the models, pass `-dry` to the annotate
worker (`go run ./cmd/annotate serve -dry`). It uses a fake offline panel.

## The commands

The API and the workers run as servers. The rest are one-off commands.

- `go run ./cmd/api` — the HTTP API the frontend uses.
- `go run ./cmd/scrape serve` — worker that pulls review pages from Steam and Reddit until they are done.
- `go run ./cmd/annotate serve` — worker that runs the LLM panel over the reviews. Needs
  `OPENROUTER_API_KEY` (or `-dry`).
- `go run ./cmd/research serve` — worker that writes a short, neutral description of a game from a web
  search, for me to review.
- `go run ./cmd/curate -per-game 50 -games <ids>` — freezes a stratified sample of reviews into a new
  population.
- `go run ./cmd/run -population <id> -prompt <id> -annotators <ids>` — creates a run with its panel and
  settings.
- `go run ./cmd/annotate enqueue -run <id>` — queues the annotation jobs for a run (the `serve` worker
  then does the work).

Scraping and game research are usually started from the frontend, not the command line.

## Project layout

```
cmd/        entry points: api, scrape, annotate, research, curate, run
internal/   the actual code: api handlers, annotate, scrape, steam, reddit, research, db (sqlc), ...
db/         migrations and the sql queries sqlc reads
frontend/   the Next.js console
```

## Notes

This is a research artifact, written and run by one person (me), so it is not meant to be a product. Some
choices come from the thesis and not from general best practice.

A few things are on purpose:

- The taxonomy and the prompts are versioned, so old runs keep pointing at the version they used.
- When I adjudicate, the panel vote at that moment is stored with my decision. This way I can always go
  back and see what the panel had said.
- The migrations are append only. To fix an old one I add a new migration instead of editing it.

## AI use

I built this with the help of an AI assistant (Claude). I used it mostly to save time on work that
followed patterns already set in the codebase, so newer features were built the same way as the existing
ones, plus some ideation, bug fixing and refactors. I directed the work, reviewed the changes, and the
design decisions and the research are mine.
