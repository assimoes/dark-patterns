import { type ReactNode } from "react";
import Link from "next/link";
import { OperationsNav } from "@/components/operations/OperationsNav";

// The operator area shell. 
export default function OperationsLayout({
    children,
}: {
    children: ReactNode;
}) {
    return (
        <main className="mx-auto max-w-3xl px-5 py-10 sm:px-8 sm:py-14">
            <header className="mb-8">
                <h1 className="text-2xl font-semibold tracking-tight text-slate-900 sm:text-3xl">
                    Operations
                </h1>
                <p className="mt-1 text-sm text-slate-500">
                    Write actions for the annotation pipeline — register games, build
                    populations, open runs, and enqueue work.{" "}
                    <Link href="/" className="font-medium text-slate-700 hover:underline">
                        Back to dashboard
                    </Link>
                </p>
            </header>

            <div className="mb-8">
                <OperationsNav />
            </div>

            {children}
        </main>
    );
}