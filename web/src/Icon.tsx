export type IconName = "home" | "compass" | "book" | "chart" | "users" | "sparkles" | "arrow" | "logout" | "check" | "download" | "shield" | "mic" | "send" | "volume" | "diamond" | "settings" | "clock" | "offline";

const paths: Record<IconName, string> = {
  home: "m3 10 9-7 9 7M5 9v12h5v-7h4v7h5V9",
  compass: "M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18Zm4 5-3 5-5 3 3-5 5-3Z",
  book: "M12 5v16M12 6C8 3 4 4 3 5v14c3-2 6-1 9 2 3-3 6-4 9-2V5c-3-2-6-1-9 1Z",
  chart: "M4 3v18h17M8 16v-5m5 5V7m5 9V4",
  users: "M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 9v-2a7 7 0 0 1 14 0v2M17 4a4 4 0 0 1 0 8m2 3a6 6 0 0 1 3 5",
  sparkles: "m12 3 2.6 6.4L21 12l-6.4 2.6L12 21l-2.6-6.4L3 12l6.4-2.6L12 3Zm7-1v4m-2-2h4",
  arrow: "M4 12h16m-6-6 6 6-6 6",
  logout: "M10 4H4v16h6m4-13 5 5-5 5M8 12h11",
  check: "m5 12 4 4L19 6",
  download: "M12 3v12m-5-5 5 5 5-5M4 16v5h16v-5",
  shield: "m12 3 8 3v6c0 5-8 9-8 9s-8-4-8-9V6l8-3Zm-4 9 3 3 5-6",
  mic: "M12 3a3 3 0 0 0-3 3v6a3 3 0 0 0 6 0V6a3 3 0 0 0-3-3Zm-6 8v1a6 6 0 0 0 12 0v-1m-6 7v3m-3 0h6",
  send: "m3 3 18 9-18 9 4-9-4-9Zm4 9h14",
  volume: "m11 4-6 5H2v6h3l6 5V4Zm4 4a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14",
  diamond: "m3 8 4-5h10l4 5-9 13L3 8Zm0 0h18M7 3l5 18 5-18",
  settings: "M12 8a4 4 0 1 0 0 8 4 4 0 0 0 0-8Zm-2-5h4l1 3 3-1 3 3-1 3 1 3-3 3-3-1-1 3h-4l-1-3-3 1-3-3 1-3-1-3 3-3 3 1 1-3Z",
  clock: "M12 3a9 9 0 1 0 0 18 9 9 0 0 0 0-18Zm0 4v5l4 2",
  offline: "M4 9a13 13 0 0 1 16 0M7 13a8 8 0 0 1 10 0m-7 4a3 3 0 0 1 4 0M3 3l18 18",
};

export function Icon({ name, className = "" }: { name: IconName; className?: string }) {
  return <svg className={`icon ${className}`} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.7" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true"><path d={paths[name]} /></svg>;
}

export function Brand() {
  return <span className="brand"><span className="brand-mark"><Icon name="diamond" /></span><span>EnglishCraft<small>LEARN. PLAY. GROW.</small></span></span>;
}
