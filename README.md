# yamex-go

## Getting Started

To run `yamex` in development you'll need to get a few things sorted besides just compiling and running the server. 
Below is a non-exhaustive list of instructions to do so.

1. [Set up a Slack workspace](https://slack.com/intl/en-ca/help/articles/206845317-Create-a-Slack-workspace)
2. [Create an app](https://slack.com/intl/en-ca/help/articles/115005265703-Create-a-bot-for-your-workspace) - Use the provided `slack_app_manifest.yml` for ease of setup.
3. [Create a Fauna Account](https://dashboard.fauna.com/accounts/register) - It's free and you can use GitHub 
4. Create a Fauna Database + Access Key
5. `cp ./config.sample.yml ./config.yml`
6. In config.yml set `FAUNA_SECRET` to the value of the Access Key from step 4


to be continued
