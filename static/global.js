

// Icl deepseek did everything I just patched it up 🥀
function background() {
    // ---------- CONFIGURATION ----------
    const darkRGB = [0, 0, 40];        // base dark blue
    const purpleRGB = [30, 0, 50];      // peak purple (still dark)
    const baseFreq = 0.0025;              // extremely low = huge patches
    const numOctaves = 1;                 // 1 is smoothest; can go to 2 for slight texture
    // -----------------------------------

    // Compute the color matrix coefficients for:
    // output = darkRGB + noise * (purpleRGB - darkRGB)
    function computeRow(low, high) {
        const slope = (high - low) / 255;
        const intercept = low / 255;
        // We put the slope on the input R (since R=G=B in grayscale)
        return [slope, 0, 0, intercept, 0];
    }

    const rowR = computeRow(darkRGB[0], purpleRGB[0]);
    const rowG = computeRow(darkRGB[1], purpleRGB[1]);
    const rowB = computeRow(darkRGB[2], purpleRGB[2]);
    const rowA = [0, 0, 0, 1, 0];  // preserve alpha

    const matrixValues = [
        ...rowR,
        ...rowG,
        ...rowB,
        ...rowA
    ].map(v => v.toFixed(6)).join(' ');

    // Random seed for unique pattern on each refresh
    const seed = Math.floor(Math.random() * 10000);

    // Create SVG filter definition
    const svgNS = 'http://www.w3.org/2000/svg';
    const svg = document.createElementNS(svgNS, 'svg');
    svg.setAttribute('style', 'position: absolute; width: 0; height: 0;');
    svg.setAttribute('focusable', 'false');

    const filter = document.createElementNS(svgNS, 'filter');
    filter.setAttribute('id', 'purpleNoise');
    filter.setAttribute('filterUnits', 'objectBoundingBox');
    filter.setAttribute('primitiveUnits', 'objectBoundingBox');
    filter.setAttribute('x', '0%');
    filter.setAttribute('y', '0%');
    filter.setAttribute('width', '100%');
    filter.setAttribute('height', '100%');

    // Turbulence – generates the smooth noise
    const turbulence = document.createElementNS(svgNS, 'feTurbulence');
    turbulence.setAttribute('type', 'fractalNoise');
    turbulence.setAttribute('baseFrequency', baseFreq);
    turbulence.setAttribute('numOctaves', numOctaves);
    turbulence.setAttribute('seed', seed);
    turbulence.setAttribute('result', 'noise');

    // Color matrix – maps grayscale to the dark blue ↔ purple range
    const colorMatrix = document.createElementNS(svgNS, 'feColorMatrix');
    colorMatrix.setAttribute('in', 'noise');
    colorMatrix.setAttribute('type', 'matrix');
    colorMatrix.setAttribute('values', matrixValues);

    filter.appendChild(turbulence);
    filter.appendChild(colorMatrix);
    svg.appendChild(filter);
    document.body.appendChild(svg);

    // Create/update the background div to use the filter
    let bgDiv = document.getElementById('bg-noise');
    if (!bgDiv) {
        bgDiv = document.createElement('div');
        bgDiv.id = 'bg-noise';
        bgDiv.style.position = 'fixed';
        bgDiv.style.inset = '0';
        bgDiv.style.zIndex = '-1';
        bgDiv.style.background = 'transparent';
        bgDiv.style.pointerEvents = 'none';
        // Set initial opacity to 0 and add transition
        bgDiv.style.opacity = '0';
        bgDiv.style.transform = 'scale(1)';
        bgDiv.style.transition = 'opacity 0.5s ease, transform 1s ease';
        document.body.appendChild(bgDiv);
    }
    bgDiv.style.filter = 'url(#purpleNoise)';

    // Fade in after a tiny delay to ensure the filter is applied
    setTimeout(() => {
        bgDiv.style.opacity = '0.6';
        bgDiv.style.transform = 'scale(1.2)';
    }, 30); 
}

document.addEventListener('DOMContentLoaded', background);