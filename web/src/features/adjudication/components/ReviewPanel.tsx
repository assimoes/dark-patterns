export function ReviewPanel({ text }: { text: string }) {
    return (
        <section aria-label="review" className="rounded-lg border bg-gray-50 p-4">
            <h2 className="mb-2 text-sm font-medium text-gray-500">the review</h2>
            <p className="whitespace-pre-wrap text-sm leading-relaxed">{text}</p>
        </section>
    );
}