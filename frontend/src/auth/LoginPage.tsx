import { useAuth } from "./AuthProvider";

export function LoginPage() {
  const { signInWithGoogle } = useAuth();

  return (
    <div
      style={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        justifyContent: "center",
        minHeight: "100vh",
        background: "#0a0a0f",
        color: "#e0e0e0",
        fontFamily: "system-ui, sans-serif",
      }}
    >
      <h1 style={{ fontSize: "2.5rem", marginBottom: "0.5rem" }}>C2Graph</h1>
      <p style={{ color: "#888", marginBottom: "2rem" }}>
        Solana Fraud Detection Graph
      </p>
      <button
        onClick={signInWithGoogle}
        style={{
          padding: "12px 24px",
          fontSize: "1rem",
          background: "#4285f4",
          color: "white",
          border: "none",
          borderRadius: "8px",
          cursor: "pointer",
        }}
      >
        Sign in with Google
      </button>
    </div>
  );
}
