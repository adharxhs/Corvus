import { Outlet } from "react-router-dom";

export function MobileLayout() {
  return (
    <div className="mobile-shell">
      <Outlet />
    </div>
  );
}
