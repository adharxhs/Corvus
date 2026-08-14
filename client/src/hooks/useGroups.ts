import { useMemo, useState } from "react";
import type { Group } from "../types/group";

export function useGroups() {
  const [groups, setGroups] = useState<Group[]>([]);

  return useMemo(
    () => ({
      groups,
      addGroup: (group: Group) => setGroups((prev) => [...prev, group]),
      setGroups,
    }),
    [groups],
  );
}
