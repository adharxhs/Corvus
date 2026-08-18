const PROFILE_PICTURE_CACHE_KEY = "corvus-profile-pictures";

interface CachedPicture {
  imageData: string;
  version: number;
  cachedAt: number;
}

export function getCachedProfilePicture(userId: string): string | null {
  try {
    const raw = localStorage.getItem(`${PROFILE_PICTURE_CACHE_KEY}_${userId}`);
    if (!raw) return null;
    const cached = JSON.parse(raw) as CachedPicture;
    return `data:image/jpeg;base64,${cached.imageData}`;
  } catch {
    return null;
  }
}

export function cacheProfilePicture(userId: string, imageData: string, version: number): void {
  try {
    const cached: CachedPicture = { imageData, version, cachedAt: Date.now() };
    localStorage.setItem(`${PROFILE_PICTURE_CACHE_KEY}_${userId}`, JSON.stringify(cached));
  } catch {
    // Ignore quota errors
  }
}

export function clearProfilePictureCache(userId: string): void {
  localStorage.removeItem(`${PROFILE_PICTURE_CACHE_KEY}_${userId}`);
}

export function clearAllProfilePictures(): void {
  const keys: string[] = [];
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i);
    if (key?.startsWith(`${PROFILE_PICTURE_CACHE_KEY}_`)) {
      keys.push(key);
    }
  }
  keys.forEach((key) => localStorage.removeItem(key));
}
