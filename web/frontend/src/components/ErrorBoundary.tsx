import type { ReactNode } from 'react';
import { Component } from 'react';
import { trackEvent } from '../utils/metrics';

type ErrorBoundaryProps = {
  children: ReactNode;
};

type ErrorBoundaryState = {
  hasError: boolean;
  error?: Error;
};

export default class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error) {
    console.error('App ErrorBoundary caught error', error);
    trackEvent('ui_error_boundary', { message: error.message, name: error.name, stack: error.stack });
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="min-h-screen flex items-center justify-center bg-[var(--bg-main)] px-6">
          <div className="max-w-md w-full bg-[var(--bg-card)] border border-[var(--border-light)] rounded-2xl p-8 shadow-lg">
            <h1 className="text-xl font-semibold text-[var(--text-main)]">应用发生错误</h1>
            <p className="text-sm text-[var(--text-muted)] mt-3">
              请刷新页面重试，或联系管理员。
            </p>
            <div className="mt-6 flex items-center gap-3">
              <button
                onClick={() => window.location.reload()}
                className="px-4 py-2 rounded-md bg-[var(--accent)] text-white text-sm font-medium hover:opacity-90"
              >
                重新加载
              </button>
              {this.state.error?.message && (
                <span className="text-xs text-[var(--text-muted)] break-all">
                  {this.state.error.message}
                </span>
              )}
            </div>
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}
