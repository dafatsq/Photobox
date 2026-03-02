const fs = require('fs');

// We are replacing the postcard (frame2) with a 2x6 Blue Strip that holds 3 photos instead of 4
const blueStripSvg = `
<svg width="600" height="1800" xmlns="http://www.w3.org/2000/svg">
  <!-- Background -->
  <rect x="0" y="0" width="600" height="1800" fill="#ADD8E6" />
  
  <!-- Fancy Border -->
  <rect x="15" y="15" width="570" height="1770" fill="none" stroke="#FFFFFF" stroke-width="8" rx="20" />
  <rect x="25" y="25" width="550" height="1750" fill="none" stroke="#00008B" stroke-width="2" stroke-dasharray="10,5" rx="10" />
  
  <!-- Transparent Cutouts (Use a mask) -->
  <mask id="cutouts_blue">
    <rect x="0" y="0" width="600" height="1800" fill="white" />
    <!-- 3 Photos, spread out more -->
    <rect x="40" y="60" width="520" height="420" fill="black" rx="10" />
    <rect x="40" y="540" width="520" height="420" fill="black" rx="10" />
    <rect x="40" y="1020" width="520" height="420" fill="black" rx="10" />
  </mask>
  
  <rect x="0" y="0" width="600" height="1800" fill="#ADD8E6" mask="url(#cutouts_blue)" />
  
  <text x="300" y="1580" font-family="Georgia, serif" font-size="50" fill="#00008B" text-anchor="middle" font-weight="bold">MEMORIES</text>
  <text x="300" y="1630" font-family="Arial, sans-serif" font-size="20" fill="white" text-anchor="middle" letter-spacing="5">EST. 2026</text>
  
  <circle cx="300" cy="1700" r="15" fill="#FFFFFF" />
  <circle cx="250" cy="1700" r="10" fill="#FFFFFF" opacity="0.7" />
  <circle cx="350" cy="1700" r="10" fill="#FFFFFF" opacity="0.7" />
</svg>
`;

fs.writeFileSync('C:\\Users\\Davu\\PhotoboxData\\frames\\frame2.svg', blueStripSvg.trim());

console.log('Blue 3-photo strip SVG generated');
