import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "../contexts/AuthContext";
import { ThemePicker } from "../components/ThemePicker";
import { changePasswordRequest } from "../services/auth";

export function SettingsPage() {
  const { user, logout } = useAuth();
  const navigate = useNavigate();
  const [currentPassword, setCurrentPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [passwordError, setPasswordError] = useState<string | null>(null);
  const [passwordSuccess, setPasswordSuccess] = useState(false);
  const [isChangingPassword, setIsChangingPassword] = useState(false);

  async function handleChangePassword(event: FormEvent) {
    event.preventDefault();
    setPasswordError(null);
    setPasswordSuccess(false);
    if (!currentPassword || !newPassword) {
      setPasswordError("Please fill out all password fields.");
      return;
    }
    if (newPassword !== confirmPassword) {
      setPasswordError("New passwords do not match.");
      return;
    }
    setIsChangingPassword(true);
    try {
      await changePasswordRequest(currentPassword, newPassword);
      setPasswordSuccess(true);
      setCurrentPassword("");
      setNewPassword("");
      setConfirmPassword("");
    } catch (err) {
      setPasswordError(err instanceof Error ? err.message : "Failed to change password.");
    } finally {
      setIsChangingPassword(false);
    }
  }

  return (
    <div className="page settings-page">
      <header className="title-bar">
        <button type="button" className="icon-button" onClick={() => navigate(-1)} aria-label="Back">
          <svg viewBox="0 0 24 24" width="24" height="24" stroke="currentColor" fill="none" strokeWidth="2">
            <path d="M19 12H5M12 19l-7-7 7-7" />
          </svg>
        </button>
        <h1 className="title-bar-name">Settings</h1>
        <span className="spacer" />
      </header>
      <section className="settings-section">
        <h3>Profile</h3>
        <p className="muted">Username: {user?.username}</p>
        <p className="muted">User ID: {user?.id}</p>
      </section>
      <section className="settings-section">
        <h3>Profile Picture</h3>
        <button type="button" className="primary-button" onClick={() => navigate("/profile-picture")}>
          Edit Profile Picture
        </button>
      </section>
      <section className="settings-section">
        <h3>Change Password</h3>
        <form className="settings-form" onSubmit={handleChangePassword}>
          <label>
            Current Password
            <input
              type="password"
              value={currentPassword}
              onChange={(e) => setCurrentPassword(e.target.value)}
              autoComplete="current-password"
            />
          </label>
          <label>
            New Password
            <input
              type="password"
              value={newPassword}
              onChange={(e) => setNewPassword(e.target.value)}
              autoComplete="new-password"
            />
          </label>
          <label>
            Confirm New Password
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              autoComplete="new-password"
            />
          </label>
          {passwordError && <p className="error-text">{passwordError}</p>}
          {passwordSuccess && <p className="success-text">Password updated successfully.</p>}
          <button type="submit" className="primary-button" disabled={isChangingPassword}>
            {isChangingPassword ? "Updating..." : "Update Password"}
          </button>
        </form>
      </section>
      <section className="settings-section">
        <ThemePicker />
      </section>
      <section className="settings-section">
        <button type="button" className="primary-button danger" onClick={() => { logout(); navigate("/login"); }}>Logout</button>
      </section>
    </div>
  );
}
