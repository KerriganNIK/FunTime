// Inline flags keep country badges consistent on systems without flag emoji support.
export function CountryFlag({ country }: { country: string }) {
  const star = <path d="m0-5 1.5 3.4 3.7.4-2.8 2.5.8 3.7L0 8l-3.2 2 .8-3.7-2.8-2.5 3.7-.4Z" />;
  let artwork;
  switch (country) {
    case 'norway': artwork = <><path fill="#bf263c" d="M0 0h30v20H0z" /><path fill="#fff" d="M8 0h6v20H8zM0 7h30v6H0z" /><path fill="#253762" d="M10 0h2v20h-2zM0 9h30v2H0z" /></>; break;
    case 'germany': artwork = <><path fill="#27272b" d="M0 0h30v7H0z" /><path fill="#cf3542" d="M0 7h30v7H0z" /><path fill="#f3c34f" d="M0 14h30v6H0z" /></>; break;
    case 'france': artwork = <><path fill="#334781" d="M0 0h10v20H0z" /><path fill="#fff" d="M10 0h10v20H10z" /><path fill="#d74851" d="M20 0h10v20H20z" /></>; break;
    case 'russia': artwork = <><path fill="#fff" d="M0 0h30v7H0z" /><path fill="#3453a0" d="M0 7h30v7H0z" /><path fill="#ce3d51" d="M0 14h30v6H0z" /></>; break;
    case 'canada': artwork = <><path fill="#fff" d="M0 0h30v20H0z" /><path fill="#d3424a" d="M0 0h7v20H0zM23 0h7v20h-7zM15 3l2 5 3-2-1 4 3 1-6 4v3h-2v-3l-6-4 3-1-1-4 3 2Z" /></>; break;
    case 'mexico': artwork = <><path fill="#fff" d="M0 0h30v20H0z" /><path fill="#337954" d="M0 0h10v20H0z" /><path fill="#cb4146" d="M20 0h10v20H20z" /><path d="m12 13 3-6 3 6Z" fill="#a68141" /><path d="M12 14h6" stroke="#337954" /></>; break;
    case 'kazakhstan': artwork = <><path fill="#48aed3" d="M0 0h30v20H0z" /><circle cx="16" cy="8" r="3" fill="#f4cf48" /><g stroke="#f4cf48">{Array.from({length:12},(_,i)=><path key={i} d="M16 3v-1" transform={`rotate(${i*30} 16 8)`} />)}<path d="m11 13 5 2 5-2" /></g><path fill="#f4cf48" d="M3 2h1v16H3z" /></>; break;
    case 'north-korea': artwork = <><path fill="#365998" d="M0 0h30v20H0z" /><path fill="#fff" d="M0 3h30v14H0z" /><path fill="#cd414b" d="M0 4h30v12H0z" /><circle cx="9" cy="10" r="4" fill="#fff" /><g fill="#cd414b" transform="translate(9 8.4) scale(.45)">{star}</g></>; break;
    case 'saudi-arabia': artwork = <><path fill="#287754" d="M0 0h30v20H0z" /><text x="15" y="10" fill="#fff" textAnchor="middle" fontSize="5" lang="ar">لا إله إلا الله</text><path d="M8 14h14l-2 1" stroke="#fff" /></>; break;
    default: artwork = <><path fill="#293e70" d="M0 0h30v20H0z" /><path d="m0 0 12 8M0 8 12 0" stroke="#fff" strokeWidth="2" /><path d="M6 0v8M0 4h12" stroke="#fff" strokeWidth="3" /><path d="M6 0v8M0 4h12" stroke="#cb4852" strokeWidth="1.5" /><g fill="#fff">{[[21,4],[24,11],[17,14],[8,14]].map(([x,y])=><g key={`${x}:${y}`} transform={`translate(${x} ${y}) scale(.25)`}>{star}</g>)}</g></>;
  }
  return <svg className="game-flag" viewBox="0 0 30 20" aria-hidden="true">{artwork}</svg>;
}
