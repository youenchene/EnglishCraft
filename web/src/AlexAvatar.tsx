import { useId } from "react";
import { AlexState } from "./tutorApi";

// Original block-world parody of Minecraft's Alex, drawn here rather than using
// an official skin, game screenshot, texture, or character asset.
export function AlexAvatar({ state, animation }: { state: AlexState; animation: boolean }) {
  const id = useId().replace(/:/g, "");
  return <svg className={`alex-avatar block-alex ${animation ? state : "still"}`} viewBox="0 0 360 420" role="img" aria-label={`Alex : ${state}`}>
    <defs>
      <linearGradient id={`${id}-shirt`} x1="0" y1="0" x2="1" y2="1"><stop stopColor="#99bb6d" /><stop offset="1" stopColor="#568047" /></linearGradient>
      <linearGradient id={`${id}-skin`} x1="0" y1="0" x2="1" y2="1"><stop stopColor="#ffdfb5" /><stop offset="1" stopColor="#e4ae87" /></linearGradient>
      <linearGradient id={`${id}-hair`} x1="0" y1="0" x2="1" y2="1"><stop stopColor="#f3a64d" /><stop offset="1" stopColor="#c56a2e" /></linearGradient>
    </defs>
    <ellipse className="alex-ground" cx="183" cy="398" rx="112" ry="12" fill="#365b37" opacity=".12" />
    <g className="alex-body">
      <path d="m100 278 23-18h123l22 18v111H100z" fill={`url(#${id}-shirt)`} />
      <path d="m246 260 22 18v111l-22-13z" fill="#416338" /><path d="M123 260h123l22 18H100z" fill="#b2cc84" />
      <path d="M153 249h54v38h-54z" fill={`url(#${id}-skin)`} /><path d="M153 249h54v12h-54z" fill="#d29a72" />
      <path d="m143 270 38 17 36-17v22h-22v13h-27v-13h-25z" fill="#759951" />
      <path d="M109 316h15v38h-15zm111-25h14v34h-14zm-87 73h14v17h-14z" fill="#b1c987" opacity=".55" />
      <path d="M100 378h146v14H100z" fill="#514a37" /><path d="M160 378h23v14h-23z" fill="#bea36a" /><path d="M164 381h15v7h-15z" fill="#645643" />
      <path d="M108 392h56v10h-56zm74 0h56v10h-56z" fill="#5c5143" />
      <path d="M81 282h31v54H81z" fill="#91b466" /><path d="m81 282 15-15h27l-11 15z" fill="#b8ce8d" /><path d="M81 325h31v12H81z" fill="#648746" />
      <g className="alex-hand">
        <path d="m82 330-23-20-8-32 27-7 12 28 18 12z" fill={`url(#${id}-skin)`} />
        <path d="M48 277v-46h12v-12h13v15h12v17h13v17H84v21H59z" fill={`url(#${id}-skin)`} />
        <path d="M48 231h12v45H48z" fill="#ffdfb9" /><path d="M72 250h12v25H72z" fill="#d9a47c" />
        <path d="M60 260h12v6H60z" fill="#c98f66" />
      </g>
      <path d="M247 282h29v69h-29z" fill="#7d9e53" /><path d="m247 282 10-15h20l-1 15z" fill="#a9c681" />
      <path d="M249 338h27v42h-27z" fill={`url(#${id}-skin)`} /><path d="M265 338h11v42h-11z" fill="#cb906a" />
      <path d="m212 326 14 12-14 13-14-13z" fill="#daf0c2" /><path d="m212 330 8 8-8 8-8-8z" fill="#8fbc8c" />
    </g>
    <g className="alex-head">
      <path className="alex-ear-aura" d="M88 157h24v51H88z" fill="#8bcaaa" opacity="0" />
      <path d="M105 96h140v164H105z" fill={`url(#${id}-skin)`} />
      <path d="m245 96 24-19v165l-24 18z" fill="#d59b74" /><path d="m105 96 24-19h140l-24 19z" fill="#ffe8c8" />
      <path d="M97 163h13v36H97zm148-6h14v39h-14z" fill="#e0aa83" />
      <path d="M100 68h148v37h-12v40h-17v-22h-25v15h-22v-13h-31v28h-27v25h-14z" fill={`url(#${id}-hair)`} />
      <path d="m100 68 25-19h148l-25 19z" fill="#ffbd65" /><path d="m248 68 25-19v106l-25 23z" fill="#bb652b" />
      <path d="M111 80h78v12h-78zm14 22h38v12h-38zm69-32h29v12h-29z" fill="#ffca75" opacity=".7" />
      <path d="M245 138h25v125h-8v26h-27v-24h-10v-26h20z" fill={`url(#${id}-hair)`} />
      <path d="M246 182h24v11h-24zm0 30h24v10h-24zm-11 47h27v10h-27z" fill="#ab5727" opacity=".65" />
      <path d="M235 280h27v9h-27z" fill="#658548" /><path d="M239 289h20v12h-20z" fill="#da8535" />
      <path className="alex-brow left" d="M129 153h35" stroke="#b66a37" strokeWidth="7" /><path className="alex-brow right" d="M193 153h34" stroke="#b66a37" strokeWidth="7" />
      <g className="alex-eyes">
        <rect x="128" y="165" width="35" height="29" rx="2" fill="#fff9e9" /><rect x="193" y="165" width="35" height="29" rx="2" fill="#fff9e9" />
        <g className="alex-pupils"><rect x="140" y="166" width="20" height="27" rx="1" fill="#648957" /><rect x="195" y="166" width="20" height="27" rx="1" fill="#648957" /><rect x="145" y="171" width="10" height="20" fill="#2b4530" /><rect x="200" y="171" width="10" height="20" fill="#2b4530" /><rect x="149" y="168" width="6" height="7" fill="white" /><rect x="204" y="168" width="6" height="7" fill="white" /></g>
      </g>
      <path d="M173 180h12v24h-12z" fill="#d99f74" /><path d="M179 202h14v5h-14z" fill="#c68d62" />
      <path d="M124 205h7v5h-7zm13 5h5v5h-5zm77 0h5v5h-5zm13-5h7v5h-7z" fill="#ca8d63" opacity=".75" />
      <path d="M117 220h12v19h-12z" fill="#f2bb99" opacity=".7" />
      <g className="alex-smile"><path d="M153 219h52v9h-8v8h-36v-8h-8z" fill="#a66251" /><path d="M159 220h39v6h-39z" fill="#fff5df" /></g>
      <g className="alex-talking-mouth"><path d="M160 213h40v24h-40z" fill="#804a3f" /><path d="M166 214h28v6h-28z" fill="#fff5df" /><path d="M169 231h22v6h-22z" fill="#dc9680" /></g>
      <path d="M114 250h120v10H114z" fill="#dca57d" opacity=".5" />
    </g>
  </svg>;
}
