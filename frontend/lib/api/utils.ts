
const API = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export class ApiError extends Error {
    constructor(
        public status: number,
        message: string,
    ) {
        super(message);
        this.name = "ApiError";
    }
}

export async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
    const res = await fetch(`${API}${path}`, {
        credentials: 'same-origin',
        headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
        ...init,
    })

    if (!res.ok) {
        const body = await res.text().catch(() => "")
        throw new ApiError(res.status, body || res.statusText)
    }

    // no content (writes essentially)
    if (res.status === 204) {
        return undefined as T
    }

    return res.json() as Promise<T>;
}