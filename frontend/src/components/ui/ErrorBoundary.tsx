import React, { Component, ErrorInfo, ReactNode } from 'react';
import { AlertOctagon, RotateCcw, Home } from 'lucide-react';
import { Button } from './Button';
import { Card } from './Card';

interface Props {
  children: ReactNode;
  fallback?: ReactNode;
}

interface State {
  hasError: boolean;
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  public state: State = {
    hasError: false,
    error: null,
  };

  public static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error };
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('Uncaught error inside component tree:', error, errorInfo);
  }

  private handleReset = () => {
    this.setState({ hasError: false, error: null });
    window.location.href = '/';
  };

  private handleReload = () => {
    window.location.reload();
  };

  public render() {
    if (this.state.hasError) {
      if (this.props.fallback) {
        return this.props.fallback;
      }

      return (
        <div className="min-h-screen flex items-center justify-center p-6 bg-background text-foreground">
          <Card className="max-w-md w-full p-8 border border-border shadow-elevated text-center space-y-5">
            <div className="w-16 h-16 rounded-2xl bg-red-500/10 text-red-600 flex items-center justify-center mx-auto border border-red-500/20">
              <AlertOctagon className="w-8 h-8" />
            </div>

            <div className="space-y-2">
              <h1 className="text-2xl font-bold font-editorial text-foreground">
                Something went wrong
              </h1>
              <p className="text-xs text-muted-foreground leading-relaxed">
                An unexpected interface error occurred. The application state has been preserved safely.
              </p>
            </div>

            {this.state.error && (
              <div className="p-3.5 rounded-lg bg-surface border border-border text-left font-mono text-[11px] text-muted-foreground overflow-x-auto max-h-28">
                {this.state.error.message}
              </div>
            )}

            <div className="pt-2 grid grid-cols-2 gap-3">
              <Button
                onClick={this.handleReload}
                variant="outline"
                size="sm"
                leftIcon={<RotateCcw className="w-3.5 h-3.5" />}
              >
                Reload
              </Button>
              <Button
                onClick={this.handleReset}
                variant="primary"
                size="sm"
                leftIcon={<Home className="w-3.5 h-3.5" />}
              >
                Home
              </Button>
            </div>
          </Card>
        </div>
      );
    }

    return this.props.children;
  }
}
