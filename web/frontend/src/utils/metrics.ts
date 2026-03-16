export function trackEvent(name: string, payload?: Record<string, unknown>) {
    if (import.meta.env.DEV) {
        console.debug('[metrics]', name, payload || {});
    }
    // TODO: integrate with Sentry / LogRocket / custom collector
}
