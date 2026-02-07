import { useState } from "react";

interface SearchBarProps {
  onScan: (address: string, depth: number) => void;
  disabled?: boolean;
}

const base58Regex = /^[1-9A-HJ-NP-Za-km-z]{32,44}$/;

export function SearchBar({ onScan, disabled }: SearchBarProps) {
  const [address, setAddress] = useState("");
  const [depth, setDepth] = useState(1);
  const [validationError, setValidationError] = useState<string | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = address.trim();
    if (!base58Regex.test(trimmed)) {
      setValidationError("Invalid Solana address");
      return;
    }
    setValidationError(null);
    onScan(trimmed, depth);
  };

  return (
    <form
      onSubmit={handleSubmit}
      style={{
        display: "flex",
        gap: "8px",
        alignItems: "center",
        padding: "16px",
        background: "#1a1a2e",
        borderBottom: "1px solid #333",
      }}
    >
      <input
        type="text"
        value={address}
        onChange={(e) => setAddress(e.target.value)}
        placeholder="Enter Solana wallet address..."
        disabled={disabled}
        style={{
          flex: 1,
          padding: "10px 14px",
          fontSize: "14px",
          fontFamily: "monospace",
          background: "#0a0a0f",
          color: "#e0e0e0",
          border: "1px solid #444",
          borderRadius: "6px",
          outline: "none",
        }}
      />
      <select
        value={depth}
        onChange={(e) => setDepth(Number(e.target.value))}
        disabled={disabled}
        style={{
          padding: "10px 14px",
          fontSize: "14px",
          background: "#0a0a0f",
          color: "#e0e0e0",
          border: "1px solid #444",
          borderRadius: "6px",
        }}
      >
        <option value={1}>1 Hop</option>
        <option value={2}>2 Hops</option>
        <option value={3}>3 Hops</option>
      </select>
      <button
        type="submit"
        disabled={disabled}
        style={{
          padding: "10px 20px",
          fontSize: "14px",
          fontWeight: "bold",
          background: disabled ? "#333" : "#4f46e5",
          color: "white",
          border: "none",
          borderRadius: "6px",
          cursor: disabled ? "not-allowed" : "pointer",
        }}
      >
        Scan
      </button>
      {validationError && (
        <span style={{ color: "#ef4444", fontSize: "13px" }}>
          {validationError}
        </span>
      )}
    </form>
  );
}
