const { generateBattleReportImage } = require('./image_generator');

async function run() {
  const battle = {
    id: 999999999,
    team_a_ids: ['p1','p2','p3','p4','p5'],
    team_b_ids: ['p6','p7','p8','p9','p10'],
    players: [
      { id: 'p1', Name: 'Alice', Equipment: { MainHand: { Tier: 8, Type: 'SWORD', Enchantment: 0, Quality: 0 } } },
      { id: 'p2', Name: 'Bob', Equipment: { MainHand: { Tier: 8, Type: 'MACE', Enchantment: 0, Quality: 0 } } },
      { id: 'p3', Name: 'Carol', Equipment: { MainHand: { Tier: 8, Type: 'BOW', Enchantment: 0, Quality: 0 } } },
      { id: 'p4', Name: 'Dave', Equipment: { MainHand: { Tier: 8, Type: 'STAFF', Enchantment: 0, Quality: 0 } } },
      { id: 'p5', Name: 'Eve', Equipment: { MainHand: { Tier: 8, Type: 'DAGGER', Enchantment: 0, Quality: 0 } } },
      { id: 'p6', Name: 'Frank', Equipment: { MainHand: { Tier: 8, Type: 'SWORD', Enchantment: 0, Quality: 0 } } },
      { id: 'p7', Name: 'Grace', Equipment: { MainHand: { Tier: 8, Type: 'MACE', Enchantment: 0, Quality: 0 } } },
      { id: 'p8', Name: 'Heidi', Equipment: { MainHand: { Tier: 8, Type: 'BOW', Enchantment: 0, Quality: 0 } } },
      { id: 'p9', Name: 'Ivan', Equipment: { MainHand: { Tier: 8, Type: 'STAFF', Enchantment: 0, Quality: 0 } } },
      { id: 'p10', Name: 'Judy', Equipment: { MainHand: { Tier: 8, Type: 'DAGGER', Enchantment: 0, Quality: 0 } } },
    ],
  };

  try {
    const out = await generateBattleReportImage(battle);
    console.log('Generated image at', out);
  } catch (err) {
    console.error('Image generation failed:', err);
  }
}

run();
