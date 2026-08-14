import { useRef, useState } from "react";
import { validateProfilePicture } from "../utils/validators";
import { ensureProfileKey, encryptProfilePicture } from "../services/tauri";
import { uploadProfilePicture } from "../services/profilePictures";
import { useWebSocket } from "../contexts/WebSocketContext";
import { useAuth } from "../contexts/AuthContext";

export function ProfilePicturePicker() {
  const inputRef = useRef<HTMLInputElement>(null);
  const { service } = useWebSocket();
  const { user } = useAuth();
  const [version, setVersion] = useState(1);
  const [busy, setBusy] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  async function handleChange() {
    const file = inputRef.current?.files?.[0];
    if (!file || !user) {
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
      await uploadProfilePicture({
        ciphertext: encrypted.ciphertext,
        nonce: encrypted.nonce,
        version,
      });
      service.send("profile_picture_updated", { version });
      setMessage("Profile picture updated");
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
      <h2>Profile Picture</h2>
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
