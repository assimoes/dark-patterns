import Link from 'next/link';
import type { ReviewItem } from '@/types/adjudication';

export function Worklist({ items, qs }: { items: ReviewItem[]; qs: string }) {
    const done = items.filter((i) => i.decided >= i.total).length;
    const next = items.find((i) => i.decided < i.total);

    return (
        <div className='space-y-4'>
            <header className='flex items-baseline justify-between'>
                <h1 className='text-xl font-semibold'>adjudication worklist</h1>
                <span className='text-sm text-gray-500'>{done}/{items.length} reviews done</span>
            </header>

            {next ? (
                <Link
                    href={`/reviews/${next.individual_id}?${qs}`}
                    className='inline-block rounded bg-black px-4 py-2 text-sm font-medium text-white'>
                    adjudicate next
                </Link>
            ) : (
                <p className='text-sm text-green-700'>all reviews done</p>
            )}

            <ul className='divide-y rounded-lg border'>
                {items.map((it) => (
                    <li
                        key={it.individual_id}
                        className="flex items-center justify-between p-3 text-sm"
                    >
                        <Link
                            href={`/reviews/${it.individual_id}?${qs}`}
                            className='font-mono hover:underline'
                        >
                            review {it.individual_id} (game {it.external_game_id})
                        </Link>
                        <span className={it.decided >= it.total ? "text-gray-400" : "font-medium text-amber-700"}>
                            {it.decided}/{it.total} patterns
                        </span>
                    </li>
                ))}
            </ul>
        </div >
    );
}
