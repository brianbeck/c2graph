import { Component, type ReactNode } from "react";

interface Props {
  children: ReactNode;
}

interface State {
  error: Error | null;
}

export class ErrorBoundary extends Component<Props, State> {
  state: State = { error: null };

  static getDerivedStateFromError(error: Error) {
    return { error };
  }

  render() {
    if (this.state.error) {
      return (
        <div
          style={{
            padding: "20px",
            background: "#1a0a0a",
            color: "#ef4444",
            fontFamily: "monospace",
            fontSize: "13px",
            whiteSpace: "pre-wrap",
            overflow: "auto",
            maxHeight: "100vh",
          }}
        >
          <h2 style={{ color: "#ef4444", margin: "0 0 12px" }}>Render Error</h2>
          <p>{this.state.error.message}</p>
          <pre style={{ color: "#888", fontSize: "11px" }}>
            {this.state.error.stack}
          </pre>
          <button
            onClick={() => this.setState({ error: null })}
            style={{
              marginTop: "12px",
              padding: "8px 16px",
              background: "#333",
              color: "#ccc",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
            }}
          >
            Retry
          </button>
        </div>
      );
    }
    return this.props.children;
  }
}
