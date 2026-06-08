import { api } from '@/lib/api';
import { Worklist } from '@/features/adjudication/components/Worklist';

export default async function WorklistPage({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | undefined>>;
}) {
  const sp = await searchParams;
  const goldRun = Number(sp.gold_run ?? 0);
  const panelRun = Number(sp.panel_run ?? 0);
  const perGame = Number(sp.per_game ?? 5);
  const taxVersion = Number(sp.tax_version ?? 1);


  if (!goldRun || !panelRun) {
    return (
      <main className='mx-auto max-w-3xl p-6 text-sm text-gray-600'>
        set the run in the URL, eg {" "}
        <code>/?gold_run=2&amp;panel_run=1&amp;per_game=5&amp;tax_version=1</code>
      </main>
    );
  }

  const items = await api.worklist({ goldRun, panelRun, perGame, taxVersion });
  const qs = `panel_run=${panelRun}&gold_run=${goldRun}&per_game=${perGame}&tax_version=${taxVersion}`;

  return (
    <main className='mx-auto max-w-3xl p-6'>
      <Worklist items={items} qs={qs} />
    </main>
  );
}