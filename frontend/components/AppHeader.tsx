"use client";

import Link from "next/link";
import { Settings2 } from "lucide-react";


export function AppHeader() {

    return (
        <div className="border-b border-slate-200/70 bg-white/70 backdrop-blur-sm">
            <div className="mx-auto flex max-w-6xl items-center justify-between px-5 py-2.5 sm:px-8">

                <div className="flex items-center gap-2">
                    <Link
                        href="/operations"
                        className="inline-flex items-center gap-1.5 rounded-lg border border-slate-200 bg-white px-2.5 py-1.5 text-xs font-medium text-slate-600 transition-colors hover:bg-slate-50"
                    >
                        <Settings2 className="size-3.5" />
                        Operations
                    </Link>
                </div>
            </div>
        </div>
    );
}