import { redirect } from "next/navigation";

// the operator area is entity-first now; land on games.
export default function OperationsIndexPage() {
    redirect("/operations/browse/games");
}
