export function CityArtwork({ role = '' }: { role?: string }) {
  return <svg className="game-city-art" viewBox="0 0 200 80" fill="none" aria-hidden="true">
    <circle cx="154" cy="24" r="17" fill="currentColor" opacity=".12" />
    <path d="M9 72h182M18 72V46h22v26m4 0V32h22v40m9 0V41h18v31m46 0V36h23v36m5 0V51h19v21" stroke="currentColor" opacity=".3" strokeWidth="2" />
    {role === 'capital' ? <path d="M94 72V36h41v36M91 36l23-20 24 20M98 41h32M103 45v22m11-22v22m11-22v22" stroke="currentColor" strokeWidth="2" />
      : role === 'military' ? <path d="M99 72V33h33v39M106 33V23h19v10M103 47h25m-25 10h25M112 23V13h20" stroke="currentColor" strokeWidth="2" />
      : role === 'security' ? <path d="M110 72s-24-15-24-38l24-9 24 9c0 23-24 38-24 38Zm-10-26 8 8 14-18" stroke="currentColor" strokeWidth="2" />
      : role === 'intelligence' ? <><path d="M109 72V40m-17 32h35M95 30l27 27c15-15 13-25 2-37L95 30Z" stroke="currentColor" strokeWidth="2" /><path d="M136 12c8 4 12 10 12 18m-12-10c4 2 6 6 6 10" stroke="currentColor" strokeWidth="2" /></>
      : <path d="M96 72V30h32v42M103 36h5m9 0h5m-19 10h5m9 0h5m-19 10h5m9 0h5M83 72V49m-8 8 8-15 8 15m40 15V49m-8 8 8-15 8 15" stroke="currentColor" strokeWidth="2" />}
  </svg>;
}
