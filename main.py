from bot import bot
import dotenv,os

dotenv.load_dotenv()
DISCORDTOKEN = os.getenv('DISCORDTOKEN')

def main():
    bot.run(DISCORDTOKEN)

if __name__ == "__main__":
    main()
