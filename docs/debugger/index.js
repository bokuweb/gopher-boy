const getPalette = c => {
  switch (c) {
    case 0:
      return [175, 197, 160, 255];
    case 1:
      return [93, 147, 66, 255];
    case 2:
      return [22, 63, 48, 255];
    case 3:
      return [0, 40, 0, 255];
  }
};

const gpuLCDC = document.querySelector(".gpu-lcdc");
const gpuSTAT = document.querySelector(".gpu-stat");
const gpuSCROLLY = document.querySelector(".gpu-scrolly");
const gpuSCROLLX = document.querySelector(".gpu-scrollx");
const gpuLY = document.querySelector(".gpu-ly");
const gpuLYC = document.querySelector(".gpu-lyc");
const gpuDMA = document.querySelector(".gpu-dma");
const gpuBGP = document.querySelector(".gpu-bgp");
const gpuOBP0 = document.querySelector(".gpu-obp0");
const gpuOBP1 = document.querySelector(".gpu-obp1");
const gpuWY = document.querySelector(".gpu-wy");
const gpuWX = document.querySelector(".gpu-wx");

const tileData0 = document.querySelector(".tiledata0");

const renderTileMap = (ctx, vram, tiles, offsetAddr, tileData0Selected) => {
  const tileMap = [];
  for (let n = 0; n < 1024; n++) {
    const tileId = vram[offsetAddr + n];
    let index = tileId;
    if (tileData0Selected) {
      index =
        (tileId & 0x80 ? new Int8Array([tileId])[0] : tileId & 0x7f) + 256;
    }
    const sprite = tiles[index];
    for (let i = 0; i < 8; i++) {
      for (let j = 0; j < 8; j++) {
        const c = getPalette(sprite[i][j]);
        const x = j + (n % 32) * 8;
        const y = i + ~~(n / 32) * 8;
        tileMap[(y * 256 + x) * 4] = c[0];
        tileMap[(y * 256 + x) * 4 + 1] = c[1];
        tileMap[(y * 256 + x) * 4 + 2] = c[2];
        tileMap[(y * 256 + x) * 4 + 3] = 255;
      }
    }
  }
  const image = ctx.createImageData(256, 256);
  image.data.set(tileMap);
  ctx.putImageData(image, 0, 0);
};

const renderWindow = (ctx, vram, tiles, offsetAddr, wx, wy) => {
  const windowMap = [];
  for (let n = 0; n < 640; n++) {
    const tileId = vram[offsetAddr + n];
    let index = tileId;
    index = (tileId & 0x80 ? new Int8Array([tileId])[0] : tileId & 0x7f) + 256;
    const sprite = tiles[index];
    for (let i = 0; i < 8; i++) {
      for (let j = 0; j < 8; j++) {
        const c = getPalette(sprite[i][j]);
        const x = j + (n % 32) * 8 + wx - 7;
        const y = i + ~~(n / 32) * 8 + wy;
        if (x >= 160 || y >= 144) {
          continue;
        }
        windowMap[(y * 160 + x) * 4] = c[0];
        windowMap[(y * 160 + x) * 4 + 1] = c[1];
        windowMap[(y * 160 + x) * 4 + 2] = c[2];
        windowMap[(y * 160 + x) * 4 + 3] = 255;
      }
    }
  }
  const image = ctx.createImageData(160, 144);
  image.data.set(windowMap);
  ctx.putImageData(image, 0, 0);
};

const getSpritePaletteID = (tileID, x, y, vram) => {
  x = x % 8;
  const addr = tileID * 0x10;
  const base = addr + y * 2;
  const l1 = vram[base];
  const l2 = vram[base + 1];
  let paletteID = 0;
  if ((l1 & (0x01 << (7 - x))) !== 0) {
    paletteID = 1;
  }
  if ((l2 & (0x01 << (7 - x))) !== 0) {
    paletteID += 2;
  }
  return paletteID;
};

const renderSprites = (ctx, oamram, vram, longSprite, obp0, obp1) => {
  const sprites = [];
  for (let i = 0; i < 40; i++) {
    const offsetY = oamram[i * 4] - 16;
    const offsetX = oamram[i * 4 + 1] - 8;
    const tileID = longSprite ? oamram[i * 4 + 2] & 0xfe : oamram[i * 4 + 2];
    const config = oamram[i * 4 + 3];
    const yFlip = (config & 0x40) !== 0;
    const xFlip = (config & 0x20) !== 0;
    const isPallette1 = config & (0x10 != 0);
    const height = longSprite ? 16 : 8;

    for (let x = 0; x < 8; x++) {
      for (let y = 0; y < height; y++) {
        if (offsetX + x < 0 || offsetX + x >= 160) {
          continue;
        }
        if (offsetY + y < 0 || offsetY + y >= 144) {
          continue;
        }
        const paletteID = getSpritePaletteID(tileID, x, y, vram);
        const adjustedX = xFlip ? 7 - x : x;
        const adjustedY = yFlip ? 7 - y : y;
        const v = isPallette1
          ? (obp1 >> (paletteID * 2)) & 0x03
          : (obp0 >> (paletteID * 2)) & 0x03;
        if (paletteID !== 0) {
          const c = getPalette(v);
          const base =
            ((offsetY + adjustedY) * 160 + (adjustedX + offsetX)) * 4;
          sprites[base] = c[0];
          sprites[base + 1] = c[1];
          sprites[base + 2] = c[2];
          sprites[base + 3] = c[3];
        }
      }
    }
  }

  const image = ctx.createImageData(160, 144);
  image.data.set(sprites);
  ctx.putImageData(image, 0, 0);
};

const buildSprite = (vram, spriteNum) => {
  const sprite = [];
  for (let y = 0; y < 8; y++) {
    for (let x = 0; x < 8; x++) {
      if (!sprite[y]) sprite[y] = [];
      let v = 0;
      if (vram[spriteNum * 16 + y * 2] & (0x80 >> x)) {
        v += 1;
      }
      if (vram[spriteNum * 16 + y * 2 + 1] & (0x80 >> x)) {
        v += 2;
      }
      sprite[y][x] = v;
    }
  }
  return sprite;
};

const createTileData = vram => {
  const imageData = [];
  const renderSprite = (sprite, spriteNum) => {
    for (let i = 0; i < 8; i++) {
      for (let j = 0; j < 8; j++) {
        const c = getPalette(sprite[i][j]);
        const x = j + (spriteNum % 16) * 8;
        const y = i + ~~(spriteNum / 16) * 8;
        imageData[(y * 256 + x) * 4] = c[0];
        imageData[(y * 256 + x) * 4 + 1] = c[1];
        imageData[(y * 256 + x) * 4 + 2] = c[2];
        imageData[(y * 256 + x) * 4 + 3] = 255;
      }
    }
  };
  const tiles = [];
  for (let i = 0; i < 384; i++) {
    const sprite = buildSprite(vram, i);
    renderSprite(sprite, i);
    tiles.push(sprite);
  }
  return { imageData, tiles };
};

const renderTileData = imageData => {
  const ctx = tileData0.getContext("2d");
  const image = ctx.createImageData(256, 256);
  image.data.set(imageData);
  ctx.putImageData(image, 0, 0);
};

export const renderDebugInfo = gb => {
  let vram = new Uint8Array(0x2000 * 4);
  let oamram = new Uint8Array(0x800 * 4);
  gb.getVRAM(vram);
  gb.getOAMRAM(oamram);

  const lcdc = gb.readGPU(0);
  const stat = gb.readGPU(1);
  const scrollY = gb.readGPU(2);
  const scrollX = gb.readGPU(3);
  const ly = gb.readGPU(4);
  const lyc = gb.readGPU(5);
  const dma = gb.readGPU(6);
  const bgp = gb.readGPU(7);
  const obp0 = gb.readGPU(8);
  const obp1 = gb.readGPU(9);
  const wy = gb.readGPU(10);
  const wx = gb.readGPU(11);

  gpuLCDC.textContent = `0x${lcdc.toString(16)}`;
  gpuSTAT.textContent = `0x${stat.toString(16)}`;
  gpuSCROLLY.textContent = `0x${scrollY.toString(16)}`;
  gpuSCROLLX.textContent = `0x${scrollX.toString(16)}`;
  gpuLY.textContent = `0x${ly.toString(16)}`;
  gpuLYC.textContent = `0x${lyc.toString(16)}`;
  gpuDMA.textContent = `0x${dma.toString(16)}`;
  gpuBGP.textContent = `0x${bgp.toString(16)}`;
  gpuOBP0.textContent = `0x${obp0.toString(16)}`;
  gpuOBP1.textContent = `0x${obp1.toString(16)}`;
  gpuWY.textContent = `0x${wy.toString(16)}`;
  gpuWX.textContent = `0x${wx.toString(16)}`;

  // TileData
  const tileData0Selected = () => (lcdc & 0x10) !== 0x10;
  const { imageData, tiles } = createTileData(vram);
  renderTileData(imageData);

  // TileMao
  const map0Ctx = document.querySelector(".tilemap0-screen").getContext("2d");
  const map1Ctx = document.querySelector(".tilemap1-screen").getContext("2d");
  renderTileMap(map0Ctx, vram, tiles, 0x1800, tileData0Selected());
  renderTileMap(map1Ctx, vram, tiles, 0x1c00, tileData0Selected());
  map0Ctx.beginPath();
  map0Ctx.rect(scrollX, scrollY, 160, 144);
  map0Ctx.strokeStyle = "rgb(0, 0, 255)";
  map0Ctx.stroke();

  // Sprites
  const spritesCtx = document.querySelector(".sprites").getContext("2d");
  renderSprites(spritesCtx, oamram, vram, !!(lcdc & 0x04), obp0, obp1);

  // Window
  const windowCtx = document.querySelector(".window").getContext("2d");
  const offset = lcdc & 0x40 ? 0x1c00 : 0x1800;
  renderWindow(windowCtx, vram, tiles, offset, wx, wy);
};
