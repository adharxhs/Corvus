import { useEffect, useState } from "react";
import { initials } from "../utils/format";
import { getGroupProfilePicture } from "../services/groups";
import { getCachedGroupPicture, cacheGroupPicture } from "../services/groupProfilePictureCache";

interface GroupAvatarProps {
  name: string;
  groupId?: string;
  size?: number;
  className?: string;
  pictureVersion?: number;
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

export function GroupAvatar({ name, groupId, size = 42, className = "", pictureVersion }: GroupAvatarProps) {
  const bg = PALETTE[hashCode(groupId || name) % PALETTE.length];
  const [src, setSrc] = useState<string | null>(() => {
    if (!groupId) return null;
    return getCachedGroupPicture(groupId);
  });

  useEffect(() => {
    if (!groupId) return;
    const cached = getCachedGroupPicture(groupId);
    if (cached && !pictureVersion) {
      setSrc(cached);
      return;
    }
    let active = true;
    getGroupProfilePicture(groupId)
      .then((pic) => {
        if (active && pic?.image_data) {
          cacheGroupPicture(groupId, pic.image_data, pic.version);
          setSrc(`data:image/jpeg;base64,${pic.image_data}`);
        }
      })
      .catch(() => {});
    return () => { active = false; };
  }, [groupId, pictureVersion]);

  return (
    <div
      className={`avatar ${className}`}
      style={{ width: size, height: size, background: bg, fontSize: Math.round(size * 0.38), color: "#FFF", overflow: "hidden" }}
    >
      {src ? (
        <img src={src} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
      ) : (
        <span>{initials(name)}</span>
      )}
    </div>
  );
}
