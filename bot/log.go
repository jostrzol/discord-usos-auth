package bot

import (
	"strings"
)

const maxMsgLen = 2000

type snippet struct {
	start int
	end   int
	lang  string
}

func fragmentMsg(text *string) []string {
	*text = strings.ReplaceAll(*text, "\t", "    ") // discord counts tabs as 4 spaces, so must replace for accurate counting

	codeSnippetSign := "```"
	codeSnippets := make([]snippet, 0)

	{
		codeSnippetEnd := -len(codeSnippetSign)
		for true {
			codeSnippetStart := strings.Index((*text)[codeSnippetEnd+len(codeSnippetSign):], codeSnippetSign)
			if codeSnippetStart == -1 {
				break
			}
			codeSnippetStart += codeSnippetEnd + len(codeSnippetSign)
			codeSnippetEnd = strings.Index((*text)[codeSnippetStart+len(codeSnippetSign):], codeSnippetSign)
			if codeSnippetEnd == -1 {
				break
			}
			codeSnippetEnd += codeSnippetStart + len(codeSnippetSign)

			langLen := strings.Index((*text)[codeSnippetStart+len(codeSnippetSign):codeSnippetEnd], "\n")
			lang := ""
			if langLen > 0 {
				lang = (*text)[codeSnippetStart+len(codeSnippetSign) : codeSnippetStart+len(codeSnippetSign)+langLen]
			}
			codeSnippets = append(codeSnippets, snippet{codeSnippetStart, codeSnippetEnd, lang})
		}
	}

	var wasInsideSnippet *snippet
	start := 0
	result := make([]string, 0, len(*text)/maxMsgLen)

	for true {
		maxFragLen := maxMsgLen
		prefix := ""
		suffix := ""

		if wasInsideSnippet != nil {
			prefix = codeSnippetSign + wasInsideSnippet.lang + "\n"
			maxFragLen -= len(prefix)
		}

		if start+maxFragLen > len(*text) {
			// reached end
			result = append(result, prefix+(*text)[start:]+suffix)
			break
		}

		// try to find a better position to cut (at newline)
		end := strings.LastIndex((*text)[start:start+maxFragLen], "\n") + start
		if end <= start {
			end = maxMsgLen + start // the line is too long, will break it
		}

		shrinkToSnippet := 0
		for i, snippet := range codeSnippets {
			if end >= snippet.start {
				// after or inside snippet start
				if end < snippet.start+len(codeSnippetSign) {
					// inside start snippet sign
					end = snippet.start - 1
					break
				}
				if end >= snippet.end {
					// after or inside snippet end
					if end < snippet.end+len(codeSnippetSign) {
						// inside end snippet sign
						end = snippet.end - 1 // move back inside snippet, shame but cannot go forward because of limit
					} else {
						shrinkToSnippet = i
						continue
					}
				}
				// inside snippet
				wasInsideSnippet = &snippet
				suffix = codeSnippetSign // must place trailing codeSnippetSign
				maxFragLen -= len(suffix)
				break
			}
		}
		codeSnippets = codeSnippets[shrinkToSnippet:]

		// recaultulate end index (bacuase maxFragLen might have been changed)
		if end-start > maxFragLen {
			newEnd := strings.LastIndex((*text)[start:end-1], "\n") + start
			if newEnd <= start {
				end = start + maxFragLen // again, the line is too long, must break it
			} else {
				end = newEnd
			}
		}

		result = append(result, prefix+(*text)[start:end]+suffix)

		start = end + 1
	}
	return result
}

// logDiscord logs a message to all log channels of a guild
func (bot *UsosBot) logDiscord(guildID string, text string) error {
	var msgs []string
	if len(text) > maxMsgLen {
		msgs = fragmentMsg(&text)
	} else {
		msgs = []string{text}
	}

	for channelID := range bot.getGuildUsosInfo(guildID).LogChannelIDs {
		channel, err := bot.getLogChannel(guildID, channelID)
		if err != nil {
			if IsNotFound(err) {
				continue
			}
			return err
		}
		for _, msg := range msgs {
			_, err = bot.ChannelMessageSend(channel.ID, msg)
			if err != nil {
				return err
			}
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
