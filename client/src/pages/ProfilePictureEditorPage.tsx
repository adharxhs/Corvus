import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { PageHeader } from "../components/PageHeader";
import { ProfilePictureCrop } from "../components/ProfilePictureCrop";
import { AppAvatar } from "../components/AppAvatar";
import { validateProfilePicture } from "../utils/validators";
import { uploadProfilePicture, getProfilePicture } from "../services/profilePictures";
import { uploadGroupProfilePicture, getGroupProfilePicture } from "../services/groups";
import { cacheProfilePicture } from "../services/profilePictureCache";
import { clearGroupPictureCache } from "../services/groupProfilePictureCache";
import { useAuth } from "../contexts/AuthContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { TYPE_PROFILE_PICTURE_UPDATED, TYPE_GROUP_PROFILE_PICTURE_UPDATED } from "../protocol/messageTypes";

type Stage = "pick" | "crop" | "preview";

export function ProfilePictureEditorPage() {
  const { user } = useAuth();
  const { service } = useWebSocket();
  const navigate = useNavigate();
  const { groupId } = useParams<{ groupId?: string }>();
  const inputRef = useRef<HTMLInputElement>(null);
  const [stage, setStage] = useState<Stage>("pick");
  const [fileSrc, setFileSrc] = useState<string | null>(null);
  const [croppedBase64, setCroppedBase64] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [version, setVersion] = useState(1);
  const isGroup = Boolean(groupId);

  useEffect(() => {
    if (!user && !isGroup) return;
    let active = true;
    const promise = isGroup && groupId
      ? getGroupProfilePicture(groupId)
      : getProfilePicture(user!.id);
    promise
      .then((pic) => { if (active) setVersion(pic.version + 1); })
      .catch(() => { if (active) setVersion(1); });
    return () => { active = false; };
  }, [user, isGroup, groupId]);

  async function handleFileSelect(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    const validationError = validateProfilePicture(file);
    if (validationError) {
      setError(validationError);
      return;
    }
    setError(null);
    const reader = new FileReader();
    reader.onload = () => {
      setFileSrc(reader.result as string);
      setStage("crop");
    };
    reader.readAsDataURL(file);
  }

  async function handleUpload() {
    if (!croppedBase64) return;
    setBusy(true);
    setError(null);
    try {
      if (isGroup && groupId) {
        await uploadGroupProfilePicture(groupId, { image_data: croppedBase64, version });
        clearGroupPictureCache(groupId);
        service.send(TYPE_GROUP_PROFILE_PICTURE_UPDATED, { group_id: groupId, version });
      } else if (user) {
        await uploadProfilePicture({ image_data: croppedBase64, version });
        cacheProfilePicture(user.id, croppedBase64, version);
        service.send(TYPE_PROFILE_PICTURE_UPDATED, { version });
      }
      navigate(-1);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Upload failed");
    } finally {
      setBusy(false);
    }
  }

  const title = isGroup ? "Group Picture" : "Profile Picture";

  return (
    <div className="page profile-editor-page">
      <PageHeader onBack={() => navigate(-1)} title={title} />
      <div className="profile-editor-content">
        {stage === "pick" && (
          <>
            <input ref={inputRef} type="file" accept="image/*" hidden onChange={handleFileSelect} />
            <button type="button" className="upload-button upload-button-large" onClick={() => inputRef.current?.click()}>
              <svg viewBox="0 0 24 24" width="32" height="32" stroke="currentColor" fill="none" strokeWidth="1.8">
                <path d="M12 16V4m0 0l-4 4m4-4l4 4" />
                <path d="M4 17v2a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1v-2" />
              </svg>
              Choose Photo
            </button>
            <p className="muted">Your picture will be shown to your contacts and chat request recipients.</p>
          </>
        )}

        {stage === "crop" && fileSrc && (
          <div className="profile-crop-stage">
            <ProfilePictureCrop src={fileSrc} onCrop={setCroppedBase64} />
            <div className="button-group-vertical">
              <button type="button" className="primary-button" onClick={() => { setStage("preview"); if (inputRef.current) inputRef.current.value = ""; }}>
                Done
              </button>
              <button type="button" className="secondary" onClick={() => { setStage("pick"); setFileSrc(null); if (inputRef.current) inputRef.current.value = ""; }}>
                Choose Different Photo
              </button>
            </div>
          </div>
        )}

        {stage === "preview" && croppedBase64 && (
          <div className="profile-preview-stage">
            <p className="muted" style={{ marginBottom: 12 }}>This is how it appears in chat:</p>
            <div className="profile-chat-demo">
              <div className="chat-demo-header">
                <AppAvatar label={user?.username ?? "You"} src={`data:image/jpeg;base64,${croppedBase64}`} size={40} />
                <span className="chat-demo-name">{isGroup ? (groupId ?? "Group") : (user?.username ?? "You")}</span>
              </div>
              <div className="chat-demo-messages">
                <div className="chat-demo-bubble in">
                  <AppAvatar label="Other" size={28} />
                  <span>Hey! How are you?</span>
                </div>
                <div className="chat-demo-bubble out">
                  <span>I'm great, thanks!</span>
                  <AppAvatar label={user?.username ?? "You"} src={`data:image/jpeg;base64,${croppedBase64}`} size={28} />
                </div>
              </div>
            </div>
            <div className="button-group-vertical" style={{ marginTop: 16 }}>
              <button type="button" className="primary-button" disabled={busy} onClick={() => void handleUpload()}>
                {busy ? "Uploading…" : "Upload"}
              </button>
              <button type="button" className="secondary" onClick={() => { setStage("crop"); if (inputRef.current) inputRef.current.value = ""; }}>
                Edit
              </button>
            </div>
          </div>
        )}

        {error && <p className="error-text">{error}</p>}
      </div>
    </div>
  );
}
