const fs = require('fs');

// Simple script to generate SVG frames and convert to PNG (since Canvas is hard to install natively on Windows without build tools)
// We will just generate SVGs and leave them as SVGs or see if Wails can load SVGs, or use a basic HTML trick.
// Wait, actually, the easiest way to generate a frame with transparent cutouts is an SVG.
const stripSvg = `
<svg width="600" height="1800" xmlns="http://www.w3.org/2000/svg">
  <!-- Background -->
  <rect x="0" y="0" width="600" height="1800" fill="#FFB6C1" />
  <!-- Border -->
  <rect x="20" y="20" width="560" height="1760" fill="none" stroke="#FFFFFF" stroke-width="10" />
  
  <!-- Transparent Cutouts (Use a mask) -->
  <mask id="cutouts">
    <rect x="0" y="0" width="600" height="1800" fill="white" />
    <rect x="40" y="40" width="520" height="350" fill="black" />
    <rect x="40" y="430" width="520" height="350" fill="black" />
    <rect x="40" y="820" width="520" height="350" fill="black" />
    <rect x="40" y="1210" width="520" height="350" fill="black" />
  </mask>
  
  <rect x="0" y="0" width="600" height="1800" fill="#FFB6C1" mask="url(#cutouts)" />
  
  <text x="300" y="1700" font-family="Arial" font-size="40" fill="white" text-anchor="middle" font-weight="bold">PHOTOBOX</text>
  <text x="300" y="1750" font-family="Arial" font-size="20" fill="white" text-anchor="middle">est. 2026</text>
</svg>
`;

fs.writeFileSync('C:\\Photobox-main\\frontend\\public\\frames\\frame1.svg', stripSvg.trim());

const postcardSvg = `
<svg width="1200" height="1800" xmlns="http://www.w3.org/2000/svg">
  <!-- Background -->
  <rect x="0" y="0" width="1200" height="1800" fill="#ADD8E6" />
  <!-- Border -->
  <rect x="30" y="30" width="1140" height="1740" fill="none" stroke="#FFFFFF" stroke-width="15" />
  
  <!-- Transparent Cutouts (Use a mask) -->
  <mask id="cutouts2">
    <rect x="0" y="0" width="1200" height="1800" fill="white" />
    <rect x="50" y="50" width="525" height="750" fill="black" />
    <rect x="625" y="50" width="525" height="750" fill="black" />
    <rect x="50" y="850" width="525" height="750" fill="black" />
    <rect x="625" y="850" width="525" height="750" fill="black" />
  </mask>
  
  <rect x="0" y="0" width="1200" height="1800" fill="#ADD8E6" mask="url(#cutouts2)" />
  
  <text x="600" y="1700" font-family="Arial" font-size="60" fill="white" text-anchor="middle" font-weight="bold">MEMORIES</text>
</svg>
`;

fs.writeFileSync('C:\\Photobox-main\\frontend\\public\\frames\\frame2.svg', postcardSvg.trim());

console.log('SVGs generated');
