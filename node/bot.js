const { Client, GatewayIntentBits, REST, Routes, SlashCommandBuilder } = require('discord.js');
require('dotenv').config();
const fetch = require('node-fetch');

const GO_SERVER = process.env.GO_SERVER || 'http://localhost:8080';
const TOKEN = process.env.DISCORD_TOKEN;
const CLIENT_ID = process.env.CLIENT_ID;

if (!TOKEN) {
  console.error('Missing DISCORD_TOKEN in environment. Create a .env file based on .env.example');
  process.exit(1);
}

const client = new Client({ intents: [GatewayIntentBits.Guilds, GatewayIntentBits.GuildMessages, GatewayIntentBits.MessageContent] });

async function registerCommands() {
  if (!CLIENT_ID) return;

  const commands = [
    new SlashCommandBuilder()
      .setName('setchannel')
      .setDescription('Sets the channel for hellgate battle reports.')
      .addStringOption((opt) => opt.setName('server').setDescription('The server to get reports from.').setRequired(true).addChoices(
        { name: 'Europe', value: 'europe' },
        { name: 'Americas', value: 'americas' },
        { name: 'Asia', value: 'asia' }
      ))
      .addStringOption((opt) => opt.setName('mode').setDescription('The hellgate mode (2v2 or 5v5).').setRequired(true).addChoices(
        { name: '5v5', value: '5v5' },
        { name: '2v2', value: '2v2' }
      ))
      .addChannelOption((opt) => opt.setName('channel').setDescription('The channel where reports will be sent.').setRequired(true)).toJSON(),
  ];

  const rest = new REST({ version: '10' }).setToken(TOKEN);
  try {
    await rest.put(Routes.applicationCommands(CLIENT_ID), { body: commands });
    console.log('Registered global commands');
  } catch (err) {
    console.error('Failed to register commands:', err);
  }
}

client.once('ready', async () => {
  console.log(`Logged in as ${client.user.tag}`);
  await registerCommands();
  startPoller();
});

client.on('interactionCreate', async (interaction) => {
  if (!interaction.isChatInputCommand()) return;
  if (interaction.commandName === 'setchannel') {
    if (!interaction.guild) {
      await interaction.reply({ content: 'This command can only be used in a server.', ephemeral: true });
      return;
    }
    const channel = interaction.options.getChannel('channel');
    if (!channel) {
      await interaction.reply({ content: 'Invalid channel.', ephemeral: true });
      return;
    }
    if (!channel.permissionsFor || !channel.permissionsFor(interaction.guild.me)?.has('SendMessages')) {
      // best-effort check; may still fail later
    }

    const server = interaction.options.getString('server');
    const mode = interaction.options.getString('mode');

    const body = { server, mode, guild_id: String(interaction.guild.id), channel_id: String(channel.id) };
    try {
      const res = await fetch(`${GO_SERVER}/channels`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) throw new Error(`status ${res.status}`);
      await interaction.reply({ content: `Hellgate ${mode} reports for **${server}** will now be sent to ${channel}.`, ephemeral: false });
    } catch (err) {
      console.error('Failed to save channel mapping:', err);
      await interaction.reply({ content: 'Failed to save channel mapping.', ephemeral: true });
    }
  }
});

async function startPoller() {
  const intervalMinutes = parseInt(process.env.BATTLE_CHECK_INTERVAL_MINUTES || '1', 10);
  setInterval(async () => {
    try {
      const res = await fetch(`${GO_SERVER}/recent_battles`);
      if (!res.ok) throw new Error(`status ${res.status}`);
      const recent = await res.json();

      // fetch configured channels
      const chRes = await fetch(`${GO_SERVER}/channels`);
      const channels = chRes.ok ? await chRes.json() : [];

      // build mapping server->mode->{guild:channel}
      const mapping = {};
      for (const c of channels) {
        mapping[c.server] = mapping[c.server] || {};
        mapping[c.server][c.mode] = mapping[c.server][c.mode] || {};
        mapping[c.server][c.mode][c.guild_id] = c.channel_id;
      }

      for (const server of Object.keys(recent)) {
        for (const mode of Object.keys(recent[server])) {
          const battles = recent[server][mode];
          if (!battles || battles.length === 0) continue;
          const cfg = mapping[server] && mapping[server][mode];
          if (!cfg) continue;

          for (const bid of Object.values(cfg)) {
            const channelId = String(bid);
            const channel = await client.channels.fetch(channelId).catch(() => null);
            if (!channel || !channel.isTextBased()) continue;
            // send simple summary placeholder; image generation will be added in next step
            for (const battle of battles) {
              try {
                const { generateBattleReportImage } = require('./image_generator');
                const imgPath = await generateBattleReportImage(battle);
                await channel.send({ files: [imgPath] }).catch((e) => console.error('send file error', e));
              } catch (err) {
                console.error('image generation/send error', err);
                const text = `New Hellgate ${mode} battle detected on ${server}: ID ${battle.id}`;
                await channel.send(text).catch((e) => console.error('send error', e));
              }
            }
          }
        }
      }
    } catch (err) {
      console.error('Poller error:', err);
    }
  }, Math.max(1, intervalMinutes) * 60 * 1000);
}

client.login(TOKEN).catch((err) => {
  console.error('Failed to login:', err);
  process.exit(1);
});
