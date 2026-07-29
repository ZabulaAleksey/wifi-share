import { useEffect, useRef } from "react";
import {
  MediaPlayer,
  MediaProvider,
  type MediaSrc,
  type MediaPlayerInstance,
} from "@vidstack/react";
import {
  DefaultAudioLayout,
  DefaultVideoLayout,
  defaultLayoutIcons,
} from "@vidstack/react/player/layouts/default";
import { X } from "lucide-react";
import type { FileItem } from "./api";

const channel = new BroadcastChannel("wifi-share-media");
const tabId = crypto.randomUUID();

type Props = {
  file: FileItem;
  source: string;
  onClose: () => void;
};

export function MediaViewer({ file, source, onClose }: Props) {
  const player = useRef<MediaPlayerInstance>(null);
  const isVideo = file.mime.startsWith("video/");

  useEffect(() => {
    const stopOtherPlayer = (event: MessageEvent<{ type: string; tabId: string }>) => {
      if (event.data.type === "playing" && event.data.tabId !== tabId) {
        player.current?.pause();
      }
    };
    channel.addEventListener("message", stopOtherPlayer);
    return () => channel.removeEventListener("message", stopOtherPlayer);
  }, []);

  return (
    <section className={`media-dock ${isVideo ? "media-dock--video" : ""}`}>
      <div className="media-dock__header">
        <div>
          <span className="eyebrow">{isVideo ? "Сейчас смотрим" : "Сейчас играет"}</span>
          <strong>{file.name}</strong>
        </div>
        <button className="icon-button" onClick={onClose} aria-label="Закрыть плеер">
          <X size={19} />
        </button>
      </div>
      <MediaPlayer
        ref={player}
        src={{ src: source, type: file.mime } as MediaSrc}
        title={file.name}
        playsInline
        onPlay={() => channel.postMessage({ type: "playing", tabId })}
      >
        <MediaProvider />
        {isVideo ? (
          <DefaultVideoLayout icons={defaultLayoutIcons} />
        ) : (
          <DefaultAudioLayout icons={defaultLayoutIcons} />
        )}
      </MediaPlayer>
    </section>
  );
}

