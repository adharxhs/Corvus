const GROUP_PICTURE_CACHE_KEY = "corvus-group-pictures";

interface CachedGroupPicture {
  imageData: string;
  version: number;
  cachedAt: number;
}

export function getCachedGroupPicture(groupId: string): string | null {
  try {
    const raw = localStorage.getItem(`${GROUP_PICTURE_CACHE_KEY}_${groupId}`);
    if (!raw) return null;
    const cached = JSON.parse(raw) as CachedGroupPicture;
    return `data:image/jpeg;base64,${cached.imageData}`;
  } catch {
    return null;
  }
}

export function cacheGroupPicture(groupId: string, imageData: string, version: number): void {
  try {
    const cached: CachedGroupPicture = { imageData, version, cachedAt: Date.now() };
    localStorage.setItem(`${GROUP_PICTURE_CACHE_KEY}_${groupId}`, JSON.stringify(cached));
  } catch {
    // Ignore quota errors
  }
}

export function clearGroupPictureCache(groupId: string): void {
  localStorage.removeItem(`${GROUP_PICTURE_CACHE_KEY}_${groupId}`);
}

export function clearAllGroupPictures(): void {
  const keys: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key?.startsWith(`${GROUP_PICTURE_CACHE_KEY}_`)) {
      keys.push(key);
    }
  }
  keys.forEach((key) => localStorage.removeItem(key));
}
