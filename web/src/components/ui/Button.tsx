import type { ButtonHTMLAttributes } from "react";

type Props = ButtonHTMLAttributes<HTMLButtonElement> & { variant?: "primary" | "secondary" };

export function Button({ variant = "primary", className = "", ...props }: Props) {
    const base = "rounded px-4 py-2 text-sm font-medium disabled: opacity-50";
    const tone =
        variant === "primary" ? "bg-black text-white hover:bg-gray-800"
            : "border border-gray-300 hover:bg-gray-50";

    return <button className={`${base} ${tone} ${className}`} {...props} />
}

