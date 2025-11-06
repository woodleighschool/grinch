import { ErrorBoundary as ReactErrorBoundary } from "react-error-boundary";
import { ErrorInfo } from "react";
import { Icons } from "./Icons";
import { Button } from "./Button";

interface ErrorFallbackProps {
  error: Error;
  resetErrorBoundary: () => void;
}

function ErrorFallback({ error, resetErrorBoundary }: ErrorFallbackProps) {
  return (
    <div className="error-boundary">
      <div className="error-content">
        <Icons.Shield className="error-icon" />
        <h2>Something went wrong</h2>
        <p className="error-message">
          We encountered an unexpected error. Please try refreshing the page or contact support if the problem persists.
        </p>
        <details className="error-details">
          <summary>Technical Details</summary>
          <pre className="error-stack">{error.message}</pre>
        </details>
        <div className="error-actions">
          <Button onClick={resetErrorBoundary} variant="primary">
            Try again
          </Button>
          <Button onClick={() => window.location.reload()} variant="secondary">
            Refresh page
          </Button>
        </div>
      </div>
    </div>
  );
}

interface ErrorBoundaryProps {
  children: React.ReactNode;
  fallback?: React.ComponentType<ErrorFallbackProps>;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
}

export function ErrorBoundary({ children, fallback = ErrorFallback, onError }: ErrorBoundaryProps) {
  return (
    <ReactErrorBoundary
      FallbackComponent={fallback}
      onError={onError}
      onReset={() => {
        // Optional: Clear any error state or navigate to a safe route
        window.location.hash = "#/dashboard";
      }}
    >
      {children}
    </ReactErrorBoundary>
  );
}
