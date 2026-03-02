const fs = require('fs');

// We need an SVG where the cutouts are ACTUALLY transparent holes.
// The easiest way in SVG is a <path> with fill-rule="evenodd".
// A large rectangle for the background, overlapping with smaller rectangles for the holes.

const stripSvg = `
<svg width="600" height="1800" xmlns="http://www.w3.org/2000/svg">
  <!-- Fancy Border -->
  <rect x="20" y="20" width="560" height="1760" fill="none" stroke="#FFFFFF" stroke-width="10" />

  <!-- Background with Holes using evenodd path -->
  <path d="
    M0,0 H600 V1800 H0 Z 
    M40,40 H560 V390 H40 Z 
    M40,430 H560 V780 H40 Z 
    M40,820 H560 V1170 H40 Z 
    M40,1210 H560 V1560 H40 Z" 
    fill="#FFB6C1" fill-rule="evenodd" 
  />

  <text x="300" y="1700" font-family="Arial" font-size="40" fill="white" text-anchor="middle" font-weight="bold">PHOTOBOX</text>
  <text x="300" y="1750" font-family="Arial" font-size="20" fill="white" text-anchor="middle">est. 2026</text>
</svg>
`;

fs.writeFileSync('C:\\Users\\Davu\\PhotoboxData\\frames\\frame1.svg', stripSvg.trim());

const blueStripSvg = `
<svg width="600" height="1800" xmlns="http://www.w3.org/2000/svg">
  <!-- Fancy Border -->
  <rect x="15" y="15" width="570" height="1770" fill="none" stroke="#FFFFFF" stroke-width="8" rx="20" />
  <rect x="25" y="25" width="550" height="1750" fill="none" stroke="#00008B" stroke-width="2" stroke-dasharray="10,5" rx="10" />

  <!-- Background with Holes using evenodd path -->
  <path d="
    M0,0 H600 V1800 H0 Z 
    M40,60 H560 V480 H40 Z 
    M40,540 H560 V960 H40 Z 
    M40,1020 H560 V1440 H40 Z" 
    fill="#ADD8E6" fill-rule="evenodd" 
  />
  
  <text x="300" y="1580" font-family="Georgia, serif" font-size="50" fill="#00008B" text-anchor="middle" font-weight="bold">MEMORIES</text>
  <text x="300" y="1630" font-family="Arial, sans-serif" font-size="20" fill="white" text-anchor="middle" letter-spacing="5">EST. 2026</text>
  
  <circle cx="300" cy="1700" r="15" fill="#FFFFFF" />
  <circle cx="250" cy="1700" r="10" fill="#FFFFFF" opacity="0.7" />
  <circle cx="350" cy="1700" r="10" fill="#FFFFFF" opacity="0.7" />
</svg>
`;

fs.writeFileSync('C:\\Users\\Davu\\PhotoboxData\\frames\\frame2.svg', blueStripSvg.trim());

console.log('Fixed SVG holes generated');
