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
  const vram = gb.getVRAM();

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

  const tileData0Selected = () => (lcdc & 0x10) !== 0x10;
  const { imageData, tiles } = createTileData(vram);
  renderTileData(imageData);
  const map0Ctx = document.querySelector(".tilemap0-screen").getContext("2d");
  const map1Ctx = document.querySelector(".tilemap1-screen").getContext("2d");
  renderTileMap(map0Ctx, vram, tiles, 0x1800, tileData0Selected());
  renderTileMap(map1Ctx, vram, tiles, 0x1c00, tileData0Selected());
  map0Ctx.beginPath();
  map0Ctx.rect(scrollX, scrollY, 160, 144);
  map0Ctx.strokeStyle = "rgb(0, 0, 255)";
  map0Ctx.stroke();
};
