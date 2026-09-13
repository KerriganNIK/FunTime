import { useId } from 'react';

export function PlanetArtwork() {
  const id = useId().replaceAll(':', '');
  return <svg viewBox="0 0 620 310" fill="none" aria-hidden="true">
    <defs>
      <clipPath id={`${id}-sphere`}><circle cx="322" cy="156" r="112" /></clipPath>
      <radialGradient id={`${id}-shade`} cx=".25" cy=".2" r=".9"><stop stopColor="#3e5532"/><stop offset="1" stopColor="#1d2e21"/></radialGradient>
    </defs>
    <path d="M82 148h55m-27-27v54M492 56l8 17 19 2-14 13 4 19-17-10-17 10 4-19-14-13 19-2z" stroke="#536d3c" strokeWidth="1.5"/>
    <circle cx="322" cy="156" r="133" stroke="#607940" strokeOpacity=".25" strokeDasharray="3 6" />
    <ellipse cx="322" cy="265" rx="102" ry="13" fill="#405932" opacity=".1"/>
    <g transform="rotate(-15 322 156)">
      <circle cx="322" cy="156" r="112" fill={`url(#${id}-shade)`}/>
      <g clipPath={`url(#${id}-sphere)`}>
        <g stroke="#d9f49d" strokeOpacity=".22" strokeWidth=".8">
          <ellipse cx="322" cy="156" rx="75" ry="112"/><ellipse cx="322" cy="156" rx="32" ry="112"/>
          <path d="M210 156h224M218 117h208M218 195h208M241 78h162M241 234h162M322 44v224"/>
        </g>
        <path d="m211 92 22-16 22 8 9 16 28 2 16 17-6 18-17 7-3 20-17 3-7-15-20-8-10-24-17-6zM284 170l25 5 16 19-4 22-16 17-7 28-13-12-5-26-12-13 2-22zM340 57l22-10 25 13 6 19-15 12-15-5-10 10-18-6-10-18zM348 107l19-10 21 12 16-8 33 15 11 25-19 13-17-10-15 10-17-19-14 8-23-10zM343 147l29-8 24 19-9 27-13 9-10 31-15-6-10-27-11-22zM403 209l22-13 22 10-6 19-22 6-18-10z" fill="#c6e884"/>
        <path d="m244 52 22-12 23 9-8 21-18 7-11-10zM310 79l9-5 6 10-7 9-9-4z" fill="#e0efa9"/>
        <circle cx="355" cy="130" r="5" fill="#fbfbed"/><circle cx="355" cy="130" r="11" stroke="#fbfbed" strokeOpacity=".5"/>
      </g>
    </g>
    <path d="M213 100C128 130 141 211 313 210c140-1 210-63 131-103" stroke="#eff7ce" strokeWidth="3"/>
    <path d="m145 159 13-12 18 5-9 7-5 13z" fill="#f8f9e9" stroke="#597643" strokeWidth="1.5"/>
    <path d="m488 213 5 12 12 5-12 5-5 12-5-12-12-5 12-5z" fill="#708d43"/>
    <circle cx="165" cy="229" r="3" fill="#708d43"/><circle cx="450" cy="55" r="3" fill="#708d43"/>
  </svg>;
}
