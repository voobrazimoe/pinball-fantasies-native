// Rasterization only. All geometry remains in the canonical SVG.
import {createRequire} from 'node:module';
import {readFileSync} from 'node:fs';
const require = createRequire(import.meta.url);
const sharp = require(process.env.PF_ICON_SHARP || '../.cache/icon-export/node_modules/sharp');
const [source, output, size] = process.argv.slice(2);
await sharp(readFileSync(source)).resize(Number(size), Number(size)).png().toFile(output);
