import { DashboardData } from "@/lib/types";
import { fetchJSON } from "@/lib/api/utils";

export default {
    dashboard: () => fetchJSON<DashboardData>('/api/dashboard'),
}