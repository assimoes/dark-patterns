
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
    const method = init?.method ?? "GET";
    const start = performance.now();

    let res: Response;
    try {
        res = await fetch(`${API}${path}`, {
            credentials: 'same-origin',
            headers: init?.body ? { 'Content-Type': 'application/json' } : undefined,
            ...init,
        })
    } catch (err) {
        // the request never reached the api (server down, cors, network). log it so a silent screen has a trail.
        console.error(`[api] ${method} ${path} failed to reach ${API}`, err);
        throw err;
    }

    const ms = Math.round(performance.now() - start);
    console.info(`[api] ${method} ${path} -> ${res.status} (${ms}ms)`);

    if (!res.ok) {
        const body = await res.text().catch(() => "")
        console.error(`[api] ${method} ${path} -> ${res.status}: ${body || res.statusText}`);
        throw new ApiError(res.status, body || res.statusText)
    }

    // no content (writes essentially)
    if (res.status === 204) {
        return undefined as T
    }

    return res.json() as Promise<T>;
}