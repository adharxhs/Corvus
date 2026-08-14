import { initials } from "../utils/format";

interface AppAvatarProps {
  label: string;
  size?: number;
  presence?: "online" | "offline" | "unknown";
  className?: string;
}

function hashCode(value: string) {
  let hash = 0;
  for (let i = 0; i < value.length; i++) {
    hash = (hash << 5) - hash + value.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash);
}

const PALETTE = ["#1E6B4A", "#5C4D9A", "#9C5A3E", "#2A6F97", "#6B2E52", "#3D6E5A", "#7A4B1F", "#48406C"];

export function AppAvatar({ label, size = 42, presence, className = "" }: AppAvatarProps) {
  const bg = PALETTE[hashCode(label) % PALETTE.length];
  return (
    <div
      className={`avatar ${className}`}
      style={{ width: size, height: size, background: bg, fontSize: Math.round(size * 0.38), color: "#FFF" }}
    >
      <span>{initials(label)}</span>
      {presence && (
        <span
          className={`presence-dot ${presence === "online" ? "presence-online" : "presence-offline"}`}
          aria-hidden
        />
      )}
    </div>
  );
}
