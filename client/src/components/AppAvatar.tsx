import { useEffect, useState } from "react";
import { initials } from "../utils/format";
import { getCachedProfilePicture, cacheProfilePicture } from "../services/profilePictureCache";
import { getProfilePicture as fetchProfilePicture } from "../services/profilePictures";

interface AppAvatarProps {
  label: string;
  src?: string;
  userId?: string;
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

export function AppAvatar({ label, src: srcProp, userId, size = 42, presence, className = "" }: AppAvatarProps) {
  const bg = PALETTE[hashCode(label) % PALETTE.length];
  const [imgSrc, setImgSrc] = useState<string | null>(() => {
    if (srcProp) return srcProp;
    return userId ? getCachedProfilePicture(userId) : null;
  });

  useEffect(() => {
    if (srcProp) {
      setImgSrc(srcProp);
      return;
    }
    if (!userId) return;
    const cached = getCachedProfilePicture(userId);
    if (cached) {
      setImgSrc(cached);
      return;
    }
    let active = true;
    fetchProfilePicture(userId)
      .then((pic) => {
        if (active && pic.image_data) {
          const url = `data:image/jpeg;base64,${pic.image_data}`;
          cacheProfilePicture(userId, pic.image_data, pic.version);
          if (active) setImgSrc(url);
        }
      })
      .catch(() => {});
    return () => { active = false; };
  }, [userId, srcProp]);

  if (imgSrc) {
    return (
      <div className={`avatar ${className}`} style={{ width: size, height: size }}>
        <img
          src={imgSrc}
          alt={label}
          style={{ width: size, height: size, borderRadius: "50%", objectFit: "cover" }}
          onError={() => setImgSrc(null)}
        />
        {presence && (
          <span className={`presence-dot ${presence === "online" ? "presence-online" : "presence-offline"}`} aria-hidden />
        )}
      </div>
    );
  }

  return (
    <div
      className={`avatar ${className}`}
      style={{ width: size, height: size, background: bg, fontSize: Math.round(size * 0.38), color: "#FFF" }}
    >
      <span>{initials(label)}</span>
      {presence && (
        <span className={`presence-dot ${presence === "online" ? "presence-online" : "presence-offline"}`} aria-hidden />
      )}
    </div>
  );
}
