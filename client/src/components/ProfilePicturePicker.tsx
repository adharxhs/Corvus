import { useEffect, useRef, useState } from "react";
import { validateProfilePicture } from "../utils/validators";
import { compressImage, uploadProfilePicture, getProfilePicture } from "../services/profilePictures";
import { cacheProfilePicture } from "../services/profilePictureCache";
import { useAuth } from "../contexts/AuthContext";

interface Feedback {
  text: string;
  error: boolean;
}

export function ProfilePicturePicker() {
  const inputRef = useRef<HTMLInputElement>(null);
  const { user } = useAuth();
  const [version, setVersion] = useState(1);
  const [busy, setBusy] = useState(false);
  const [feedback, setFeedback] = useState<Feedback | null>(null);
  const [loadingVersion, setLoadingVersion] = useState(false);

  useEffect(() => {
    if (!user) return;
    let active = true;
    setLoadingVersion(true);
    getProfilePicture(user.id)
      .then((pic) => {
        if (active) setVersion(pic.version + 1);
      })
      .catch(() => {})
      .finally(() => {
        if (active) setLoadingVersion(false);
      });
    return () => { active = false; };
  }, [user]);

  async function handleChange() {
    const file = inputRef.current?.files?.[0];
    if (!file || !user) return;

    const validationError = validateProfilePicture(file);
    if (validationError) {
      setFeedback({ text: validationError, error: true });
      if (inputRef.current) inputRef.current.value = "";
      return;
    }

    setBusy(true);
    setFeedback(null);
    try {
      const imageData = await compressImage(file, 400, 400, 0.8);
      await uploadProfilePicture({ image_data: imageData, version });
      cacheProfilePicture(user.id, imageData, version);
      setFeedback({ text: "Profile picture updated", error: false });
    } catch (err) {
      setFeedback({ text: err instanceof Error ? err.message : "Upload failed", error: true });
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="profile-picture">
      <h2>Profile Picture</h2>
      <input ref={inputRef} type="file" accept="image/*" hidden onChange={handleChange} />
      <button type="button" className="upload-button" disabled={busy || loadingVersion} onClick={() => inputRef.current?.click()}>
        {busy ? "Uploading…" : "Choose & upload photo"}
      </button>
      {feedback && <p className={feedback.error ? "error-text" : "success-text"}>{feedback.text}</p>}
    </div>
  );
}
