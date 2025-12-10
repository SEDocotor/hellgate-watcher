const { createCanvas, loadImage, registerFont } = require('canvas');
const fs = require('fs');
const path = require('path');

const BATTLE_REPORT_DIR = path.join(__dirname, '..', 'images', 'battle_reports');
fs.mkdirSync(BATTLE_REPORT_DIR, { recursive: true });

async function getItemImageUrl(item) {
  // item expected to have Tier, Type, Enchantment, Quality fields
  if (!item) return null;
  const tier = item.Tier || item.tier || 0;
  const type = item.Type || item.type || '';
  const enchant = item.Enchantment || item.enchantment || 0;
  const quality = item.Quality || item.quality || 0;
  // Render API path example used in Python
  const url = `https://render.albiononline.com/v1/item/T${tier}_${type}@${enchant}.png?count=1&quality=${enchant}`;
  return url;
}

async function downloadImage(url) {
  try {
    const res = await fetch(url);
    if (!res.ok) return null;
    const buf = Buffer.from(await res.arrayBuffer());
    return buf;
  } catch (e) {
    return null;
  }
}

async function generateBattleReportImage(battle) {
  // minimal layout: horizontally list team A and team B equipment placeholder images and names
  const width = 1200;
  const height = 600;
  const canvas = createCanvas(width, height);
  const ctx = canvas.getContext('2d');

  // background
  ctx.fillStyle = '#282828';
  ctx.fillRect(0, 0, width, height);

  // Title
  ctx.fillStyle = '#ffffff';
  ctx.font = '28px Sans';
  ctx.fillText(`Battle ${battle.id}`, 20, 40);

  // Teams
  const leftX = 50;
  const rightX = width / 2 + 50;
  const startY = 80;
  const gapY = 100;

  const drawPlayer = async (p, x, y) => {
    ctx.fillStyle = '#ffffff';
    ctx.font = '20px Sans';
    ctx.fillText(p.Name || p.name || 'Unknown', x, y);
    // attempt to load mainhand image
    const mainhand = p.Equipment && (p.Equipment.MainHand || p.Equipment.mainhand);
    const url = await getItemImageUrl(mainhand);
    if (url) {
      try {
        const buf = await downloadImage(url);
        if (buf) {
          const img = await loadImage(buf);
          ctx.drawImage(img, x, y + 10, 80, 80);
        }
      } catch (e) {
        // ignore
      }
    }
  };

  for (let i = 0; i < (battle.team_a_ids || []).length; i++) {
    const id = battle.team_a_ids[i] || battle.teamAIDs && battle.teamAIDs[i];
    const player = (battle.players || []).find(p => p.Id === id || p.id === id || p.ID === id);
    if (player) await drawPlayer(player, leftX + i * 100, startY);
  }
  for (let i = 0; i < (battle.team_b_ids || []).length; i++) {
    const id = battle.team_b_ids[i] || battle.teamBIDs && battle.teamBIDs[i];
    const player = (battle.players || []).find(p => p.Id === id || p.id === id || p.ID === id);
    if (player) await drawPlayer(player, rightX + i * 100, startY);
  }

  const filename = `battle_report_${battle.id}.png`;
  const outPath = path.join(BATTLE_REPORT_DIR, filename);
  const out = fs.createWriteStream(outPath);
  const stream = canvas.createPNGStream();
  await new Promise((resolve, reject) => {
    stream.pipe(out);
    out.on('finish', resolve);
    out.on('error', reject);
  });
  return outPath;
}

module.exports = { generateBattleReportImage };
