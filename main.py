from src.bot import bot
from src.hellgate_watcher import gen_battle_report_by_id
import dotenv
import os

dotenv.load_dotenv()
DISCORDTOKEN = os.getenv('DISCORDTOKEN')

def main():
    bot.run(DISCORDTOKEN)

if __name__ == "__main__":
    main()
