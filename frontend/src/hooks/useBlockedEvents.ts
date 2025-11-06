import { useEffect, useState } from "react";
import { BlockedEvent, listBlocked, subscribeBlockedEvents } from "../api";

export function useBlockedEvents() {
    const [events, setEvents] = useState<BlockedEvent[]>([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);

    useEffect(() => {
        let active = true;
        (async () => {
            try {
                const initial = await listBlocked();
                if (active) {
                    setEvents(Array.isArray(initial) ? initial : []);
                }
            } catch (err) {
                if (err instanceof Error) {
                    setError(err.message);
                }
            } finally {
                if (active) {
                    setLoading(false);
                }
            }
        })();

        const unsubscribe = subscribeBlockedEvents((event) => {
            if (!event || typeof event !== "object" || event.id == null) {
                return;
            }
            setEvents((current) => {
                const next = [
                    event,
                    ...(Array.isArray(current) ? current : []),
                ];
                const deduped: BlockedEvent[] = [];
                const seen = new Set<number>();
                for (const item of next) {
                    if (!seen.has(item.id)) {
                        deduped.push(item);
                        seen.add(item.id);
                    }
                    if (deduped.length >= 100) {
                        break;
                    }
                }
                return deduped;
            });
        });

        return () => {
            active = false;
            unsubscribe();
        };
    }, []);

    return { events, loading, error };
}
