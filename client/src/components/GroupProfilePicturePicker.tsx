import { useRef, useState } from "react";
import { validateProfilePicture } from "../utils/validators";
import { ensureProfileKey, encryptProfilePicture } from "../services/tauri";
import { uploadGroupProfilePicture } from "../services/groups";
import { useWebSocket } from "../contexts/WebSocketContext";

interface GroupProfilePicturePickerProps {
  groupId: string;
}

export function GroupProfilePicturePicker({ groupId }: GroupProfilePicturePickerProps) {
  const inputRef = useRef<HTMLInputElement>(null);
  const { service } = useWebSocket();
  const [version, setVersion] = useState(1);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  async function handleChange() {
    const file = inputRef.current?.files?.[0];
    if (!file) {
      return;
    }

    const validationError = validateProfilePicture(file);
    if (validationError) {
      setMessage(validationError);
      return;
    }

    setBusy(true);
    setMessage(null);

    try {
      await ensureProfileKey();
      const bytes = new Uint8Array(await file.arrayBuffer());
      const encrypted = await encryptProfilePicture(bytes);
      await uploadGroupProfilePicture(groupId, {
        ciphertext: encrypted.ciphertext,
        nonce: encrypted.nonce,
        version,
      });
      service.send("group_profile_picture_updated", { group_id: groupId, version });
      setMessage("Group picture updated");
      setVersion((prev) => prev + 1);
    } catch (error) {
      setMessage(error instanceof Error ? error.message : "Upload failed");
    } finally {
      setBusy(false);
      if (inputRef.current) {
        inputRef.current.value = "";
      }
    }
  }

  return (
    <section className="panel profile-picture">
      <h2>Group Picture</h2>
      <p className="muted">Version: {version}</p>
      <input
        ref={inputRef}
        type="file"
        accept="image/*"
        disabled={busy}
        onChange={handleChange}
      />
      {message ? <p className="muted">{message}</p> : null}
    </section>
  );
}
