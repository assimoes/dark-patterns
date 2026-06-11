"use client";

import { Card } from "@/components/ui/Card";
import { PopulationsCard } from "@/components/dashboard/PopulationsCard";
import { RunsCard } from "@/components/dashboard/RunsCard";
import { ReviewsCard } from "@/components/dashboard/ReviewsCard";
import { useDashboard } from "@/hooks/useDashboard";

// The dashboard reads its data from the Go API via
// useDashboard() (React Query).
// It owns the three top-level states — pending, error, loaded — and only renders
// the cards once `data` is in hand, so each card can take plain typed props and
// never worry about loading.
export default function DashboardPage() {
  const { data, isPending, isError, refetch } = useDashboard();

  return (
    <main className="mx-auto max-w-6xl px-5 py-10 sm:px-8 sm:py-14">
      <header className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <div className="mb-2 flex items-center gap-2">
            <span className="grid size-8 place-items-center rounded-lg bg-slate-900 text-sm font-bold text-white">
              p2
            </span>
            <span className="inline-flex items-center gap-1.5 rounded-full border border-emerald-200 bg-emerald-50 px-2.5 py-1 text-xs font-medium text-emerald-700">
              <span className="size-1.5 rounded-full bg-emerald-500" />
              live data
            </span>
          </div>
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
            Annotation dashboard
          </h1>
          <p className="mt-1 text-sm text-slate-500">
            Dark-patterns detection pipeline — corpus, runs, and panel coverage at a glance.
          </p>
        </div>

        <div className="flex items-center gap-2">
          <a
            href="/blind"
            className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-3.5 py-2 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50"
          >
            Blind labelling →
          </a>
          <a
            href="/adjudicate"
            className="inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
          >
            Open adjudication →
          </a>
        </div>
      </header>

      {isPending ? (
        <DashboardSkeleton />
      ) : isError ? (
        <DashboardError onRetry={() => refetch()} />
      ) : (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <PopulationsCard games={data.games} populations={data.populations} />
          <RunsCard runs={data.runs} />
          <div className="lg:col-span-2">
            <ReviewsCard games={data.games} reviewStats={data.reviewStats} />
          </div>
        </div>
      )}

      <footer className="mt-10 text-center text-xs text-slate-400">
        Live UI · Next.js + Tailwind · data from the dsr API
      </footer>
    </main>
  );
}

// While the first fetch is in flight, three placeholder cards keep the layout
// stable so nothing jumps when the data arrives.
function DashboardSkeleton() {
  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
      <SkeletonCard />
      <SkeletonCard />
      <div className="lg:col-span-2">
        <SkeletonCard />
      </div>
    </div>
  );
}

function SkeletonCard() {
  return (
    <div className="h-64 animate-pulse rounded-2xl border border-slate-200/80 bg-white/60 shadow-sm ring-1 ring-slate-900/[0.02]">
      <div className="border-b border-slate-100 px-6 py-5">
        <div className="flex items-center gap-3">
          <span className="size-10 rounded-xl bg-slate-100" />
          <div className="space-y-2">
            <div className="h-3 w-28 rounded bg-slate-100" />
            <div className="h-2.5 w-40 rounded bg-slate-100" />
          </div>
        </div>
      </div>
      <div className="space-y-3 px-6 py-5">
        <div className="h-3 w-full rounded bg-slate-100" />
        <div className="h-3 w-5/6 rounded bg-slate-100" />
        <div className="h-3 w-2/3 rounded bg-slate-100" />
      </div>
    </div>
  );
}

// On a failed fetch we keep the same Card shell and give it a retry, so an outage
// reads as a recoverable state rather than a broken page.
function DashboardError({ onRetry }: { onRetry: () => void }) {
  return (
    <Card
      icon={<span className="text-lg">!</span>}
      iconClass="bg-rose-50 text-rose-600"
      title="Couldn't load the dashboard"
      subtitle="The API request failed."
    >
      <p className="text-sm text-slate-500">
        Check that the backend is running, then try again.
      </p>
      <button
        type="button"
        onClick={onRetry}
        className="mt-4 inline-flex items-center gap-1.5 rounded-lg bg-slate-900 px-3.5 py-2 text-sm font-medium text-white transition-colors hover:bg-slate-700"
      >
        Retry
      </button>
    </Card>
  );
}