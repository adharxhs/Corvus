import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { getGroupProfilePicture } from "../services/groups";
import { GroupAvatar } from "./GroupAvatar";

interface GroupProfilePicturePickerProps {
  groupId: string;
  pictureVersion?: number;
}

export function GroupProfilePicturePicker({ groupId, pictureVersion }: GroupProfilePicturePickerProps) {
  const navigate = useNavigate();
  const [hasPicture, setHasPicture] = useState(false);

  useEffect(() => {
    let active = true;
    getGroupProfilePicture(groupId)
      .then(() => { if (active) setHasPicture(true); })
      .catch(() => {});
    return () => { active = false; };
  }, [groupId, pictureVersion]);

  return (
    <div className="profile-picture">
      <h2>Group Picture</h2>
      <div style={{ display: "flex", alignItems: "center", gap: 12, marginBottom: 12 }}>
        <GroupAvatar name="Group" groupId={groupId} size={64} pictureVersion={pictureVersion} />
        {hasPicture ? <span className="muted">Picture set</span> : <span className="muted">No picture yet</span>}
      </div>
      <button type="button" className="primary-button" onClick={() => navigate(`/profile-picture/group/${groupId}`)}>
        {hasPicture ? "Change Picture" : "Set Picture"}
      </button>
    </div>
  );
}
