import * as React from "react";
import { cn } from "../../lib/utils";
import { Camera, Image as ImageIcon, X } from "lucide-react";

export interface PhotoAttachmentPickerProps {
  images?: string[];
  onAddImage?: (file: File) => void;
  onRemoveImage?: (index: number) => void;
  className?: string;
}

export function PhotoAttachmentPicker({
  images = [],
  onAddImage,
  onRemoveImage,
  className,
}: PhotoAttachmentPickerProps) {
  const [showSourceDialog, setShowSourceDialog] = React.useState(false);
  const fileInputRef = React.useRef<HTMLInputElement>(null);

  function handleFileSelected(e: React.ChangeEvent<HTMLInputElement>) {
    if (e.target.files && e.target.files[0] && onAddImage) {
      onAddImage(e.target.files[0]);
    }
    setShowSourceDialog(false);
  }

  return (
    <div className={cn("w-full py-2", className)}>
      <input
        ref={fileInputRef}
        type="file"
        accept="image/*"
        className="hidden"
        onChange={handleFileSelected}
      />

      <div className="flex flex-wrap gap-2 mb-2">
        {images.map((imgUrl, idx) => (
          <div key={idx} className="relative w-14 h-14 rounded-xl overflow-hidden border border-[#e8e8ec] shadow-sm">
            <img src={imgUrl} alt="Hóa đơn" className="w-full h-full object-cover" />
            {onRemoveImage && (
              <button
                type="button"
                onClick={() => onRemoveImage(idx)}
                className="absolute top-1 right-1 w-4 h-4 rounded-full bg-[#ff5a66] text-white flex items-center justify-center text-[10px]"
              >
                <X className="w-2.5 h-2.5" />
              </button>
            )}
          </div>
        ))}
      </div>

      <button
        type="button"
        onClick={() => setShowSourceDialog(true)}
        className="w-full py-3 px-4 rounded-2xl bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] font-semibold text-sm flex items-center justify-center gap-2 active:scale-98 transition-all outline-none"
      >
        <Camera className="w-4 h-4 text-[#29495a]" />
        <span>Thêm Hình Ảnh</span>
      </button>

      {/* Floating Dialog: Photo Source (IMG_7709) */}
      {showSourceDialog && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-xs p-4 animate-in fade-in"
          onClick={() => setShowSourceDialog(false)}
        >
          <div
            className="w-full max-w-[320px] rounded-3xl bg-white p-4 shadow-2xl flex flex-col gap-2.5 select-none"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="w-full h-12 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] font-semibold text-sm flex items-center justify-center gap-2 active:scale-98 transition-all"
            >
              <ImageIcon className="w-4 h-4 text-[#29495a]" />
              <span>Mở Thư Viện Ảnh</span>
            </button>

            <button
              type="button"
              onClick={() => fileInputRef.current?.click()}
              className="w-full h-12 rounded-full bg-[#eef0f4] hover:bg-[#e2e4e9] text-[#111111] font-semibold text-sm flex items-center justify-center gap-2 active:scale-98 transition-all"
            >
              <Camera className="w-4 h-4 text-[#29495a]" />
              <span>Mở Máy ảnh</span>
            </button>

            <button
              type="button"
              onClick={() => setShowSourceDialog(false)}
              className="w-full h-12 rounded-full bg-white border border-[#e8e8ec] hover:bg-[#f8f9fa] text-[#8e8e93] font-semibold text-sm active:scale-98 transition-all"
            >
              Huỷ
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
