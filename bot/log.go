package bot

// logDiscord logs a message to all log channels of a guild
func (bot *UsosBot) logDiscord(guildID string, text string) error {
	for channelID := range bot.getGuildUsosInfo(guildID).LogChannelIDs {
		channel, err := bot.getLogChannel(guildID, channelID)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			return err
		}
		_, err = bot.ChannelMessageSend(channel.ID, text)
		if err != nil {
			return err
		}
	}
	return nil
}

// addLogChannel adds a channel to log to authorization data from the guild
func (bot *UsosBot) addLogChannel(guildID string, channelID string) error {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if guildInfo.LogChannelIDs[channelID] {
		return newErrLogChannelPresent(channelID, guildID)
	}

	guildChannels, err := bot.GuildChannels(guildID)
	if err != nil {
		return err
	}

	// only add this guild's channels
	for _, guildChannel := range guildChannels {
		if guildChannel.ID == channelID {
			guildInfo.LogChannelIDs[channelID] = true
			return nil
		}
	}

	return newErrLogChannelNotFound(channelID, guildID)
}

// removeLogChannel removes a log channel
func (bot *UsosBot) removeLogChannel(guildID string, channelID string) error {
	guildInfo := bot.getGuildUsosInfo(guildID)
	if !guildInfo.LogChannelIDs[channelID] {
		return newErrLogChannelNotFound(channelID, guildID)
	}
	delete(guildInfo.LogChannelIDs, channelID)
	return nil
}
