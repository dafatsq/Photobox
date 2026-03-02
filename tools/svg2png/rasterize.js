const sharp = require('sharp');
const fs = require('fs');
const path = require('path');

const framesDir = 'C:\\Users\\Davu\\PhotoboxData\\frames';

async function convertSVGtoPNG(filename) {
    const inputPath = path.join(framesDir, filename + '.svg');
    const outputPath = path.join(framesDir, filename + '.png');

    if (fs.existsSync(inputPath)) {
        await sharp(inputPath, { density: 300 })
            .toFormat('png')
            .toFile(outputPath);
        console.log(`Converted ${filename}.svg -> ${filename}.png`);
    } else {
        console.error(`Missing ${inputPath}`);
    }
}

async function run() {
    await convertSVGtoPNG('frame1');
    await convertSVGtoPNG('frame2');
}

run();
