import { Navigate, createHashRouter } from "react-router-dom";
import { MobileLayout } from "./layouts/MobileLayout";
import { LoginPage } from "./pages/LoginPage";
import { RegisterPage } from "./pages/RegisterPage";
import { ChatsPage } from "./pages/ChatsPage";
import { ChatScreen } from "./pages/ChatScreen";
import { ContactsPage } from "./pages/ContactsPage";
import { GroupsPage } from "./pages/GroupsPage";
import { SettingsPage } from "./pages/SettingsPage";
import { ProfilePictureEditorPage } from "./pages/ProfilePictureEditorPage";
import { OfflinePage } from "./pages/OfflinePage";
import { useAuth } from "./contexts/AuthContext";
import { useServerReachable } from "./utils/connectivity";

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  if (status !== "authenticated") return <Navigate to="/login" replace />;
  return children;
}

function RootRedirect() {
  const { status } = useAuth();
  return <Navigate to={status === "authenticated" ? "/chats" : "/login"} replace />;
}

function OfflineGuard({ children }: { children: React.ReactNode }) {
  const { status } = useAuth();
  const { reachable, retry } = useServerReachable(status !== "authenticated");

  if (status === "authenticated") return <Navigate to="/chats" replace />;

  return (
    <>
      {reachable === "offline" && (
        <div className="offline-banner">
          Server unreachable — you can still try to sign in.
          <button type="button" onClick={retry} style={{ marginLeft: 8, background: "none", border: "1px solid currentColor", borderRadius: 6, padding: "2px 8px", color: "inherit", cursor: "pointer" }}>
            Retry
          </button>
        </div>
      )}
      {children}
    </>
  );
}

export const router = createHashRouter([
  { path: "/", element: <RootRedirect /> },
  { path: "/offline", element: <OfflinePage /> },
  { path: "/login", element: <OfflineGuard><LoginPage /></OfflineGuard> },
  { path: "/register", element: <OfflineGuard><RegisterPage /></OfflineGuard> },
  {
    element: (
      <ProtectedRoute>
        <MobileLayout />
      </ProtectedRoute>
    ),
    children: [
      { path: "/chats", element: <ChatsPage /> },
      { path: "/chat/:kind/:id", element: <ChatScreen /> },
      { path: "/contacts", element: <ContactsPage /> },
      { path: "/groups", element: <GroupsPage /> },
      { path: "/settings", element: <SettingsPage /> },
      { path: "/profile-picture", element: <ProfilePictureEditorPage /> },
      { path: "/profile-picture/group/:groupId", element: <ProfilePictureEditorPage /> },
    ],
  },
]);
