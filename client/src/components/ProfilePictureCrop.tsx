import { useEffect, useRef, useState, useCallback } from "react";

interface ProfilePictureCropProps {
  src: string;
  viewportSize?: number;
  outputSize?: number;
  quality?: number;
  onCrop: (base64: string) => void;
}

interface Offset {
  x: number;
  y: number;
}

export function ProfilePictureCrop({ src, viewportSize = 240, outputSize = 400, quality = 0.82, onCrop }: ProfilePictureCropProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [img, setImg] = useState<{ width: number; height: number } | null>(null);
  const [zoom, setZoom] = useState(1);
  const [offset, setOffset] = useState<Offset>({ x: 0, y: 0 });
  const dragRef = useRef<{ startX: number; startY: number; offsetX: number; offsetY: number } | null>(null);

  useEffect(() => {
    const image = new Image();
    image.onload = () => {
      setImg({ width: image.naturalWidth, height: image.naturalHeight });
      setZoom(1);
      setOffset({ x: 0, y: 0 });
    };
    image.src = src;
  }, [src]);

  const coverScale = img ? Math.max(viewportSize / img.width, viewportSize / img.height) : 1;
  const effectiveScale = coverScale * zoom;
  const drawnWidth = img ? img.width * effectiveScale : viewportSize;
  const drawnHeight = img ? img.height * effectiveScale : viewportSize;
  const drawnLeft = (viewportSize - drawnWidth) / 2 + offset.x;
  const drawnTop = (viewportSize - drawnHeight) / 2 + offset.y;

  const maxPanX = Math.max(0, (drawnWidth - viewportSize) / 2);
  const maxPanY = Math.max(0, (drawnHeight - viewportSize) / 2);

  function clampOffset(next: Offset): Offset {
    return {
      x: Math.max(-maxPanX, Math.min(maxPanX, next.x)),
      y: Math.max(-maxPanY, Math.min(maxPanY, next.y)),
    };
  }

  function onPointerDown(e: React.PointerEvent<HTMLDivElement>) {
    (e.currentTarget as HTMLDivElement).setPointerCapture(e.pointerId);
    dragRef.current = { startX: e.clientX, startY: e.clientY, offsetX: offset.x, offsetY: offset.y };
  }

  function onPointerMove(e: React.PointerEvent<HTMLDivElement>) {
    const drag = dragRef.current;
    if (!drag) return;
    setOffset(clampOffset({ x: drag.offsetX + (e.clientX - drag.startX), y: drag.offsetY + (e.clientY - drag.startY) }));
  }

  function onPointerUp() {
    dragRef.current = null;
  }

  const generate = useCallback(() => {
    if (!img) return;
    const cropSize = viewportSize / effectiveScale;
    const srcX = (Math.max(0, drawnLeft) - drawnLeft) / effectiveScale;
    const srcY = (Math.max(0, drawnTop) - drawnTop) / effectiveScale;
    const canvas = document.createElement("canvas");
    canvas.width = outputSize;
    canvas.height = outputSize;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;
    ctx.drawImage(
      document.querySelector<HTMLImageElement>(".profile-crop-img")!,
      srcX, srcY, cropSize, cropSize,
      0, 0, outputSize, outputSize,
    );
    onCrop(canvas.toDataURL("image/jpeg", quality).split(",")[1]);
  }, [img, viewportSize, effectiveScale, drawnLeft, drawnTop, outputSize, quality, onCrop]);

  useEffect(() => {
    generate();
  }, [generate]);

  return (
    <div className="profile-crop">
      <div
        ref={containerRef}
        className="profile-crop-viewport"
        style={{ width: viewportSize, height: viewportSize }}
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        onPointerCancel={onPointerUp}
      >
        {img && (
          <img
            className="profile-crop-img"
            src={src}
            alt="Crop preview"
            draggable={false}
            style={{
              position: "absolute",
              left: 0,
              top: 0,
              width: drawnWidth,
              height: drawnHeight,
              transform: `translate(${drawnLeft}px, ${drawnTop}px)`,
              maxWidth: "none",
            }}
          />
        )}
      </div>
      <label className="profile-crop-zoom">
        <span className="muted">Zoom</span>
        <input
          type="range"
          min={1}
          max={3}
          step={0.05}
          value={zoom}
          onChange={(e) => setZoom(parseFloat(e.target.value))}
        />
      </label>
    </div>
  );
}
