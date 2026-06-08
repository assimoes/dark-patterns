INSERT INTO taxonomy_meso_levels
    (parent_id, code, name, description, version, examples, counter_examples, gray_mapping, source_mapping)
SELECT p.id, v.code, v.name, v.description, 2, v.examples, v.counter_examples, v.gray_mapping, v.source_mapping
FROM (VALUES
    ('Temporal Manipulation', 'TM-1', 'Artificial Extension',
     $$Mechanics that inflate the time required to progress or complete content far beyond what the core gameplay warrants, to boost engagement metrics or drive monetization.$$,
     ARRAY[
        $$I had to kill the same five enemies for two hours to get one upgrade — pure busywork to inflate playtime.$$,
        $$Levelling past 60 is a slog; everything past that point gives you 1% of the XP you actually need.$$],
     ARRAY[
        $$The endgame is challenging and requires patience, but I never felt my time was being padded artificially.$$],
     ARRAY[$$Grinding$$, $$Gamification$$],
     ARRAY[$$Zagal: Grinding$$, $$DPGames: Grinding, Infinite Treadmill$$, $$Petrovskaya: Paying to Reduce Grinding/Waiting$$]),

    ('Temporal Manipulation', 'TM-2', 'Scheduled Coercion',
     $$Mechanics that require or pressure the player to play at specific times or intervals dictated by the game, rather than at the player's convenience.$$,
     ARRAY[
        $$Miss a daily login and your streak resets — you lose the weekly bonus and three days of progress.$$,
        $$Raids only run between 8pm and 10pm, so if you have a real life you're locked out of the best loot.$$],
     ARRAY[
        $$There's a weekly tournament every Saturday, but you can join any week — no penalty for skipping.$$],
     ARRAY[$$Nagging$$, $$Urgency$$],
     ARRAY[$$Zagal: Playing by Appointment, Daily Rewards$$, $$DPGames: Playing by Appointment, Daily Rewards$$]),

    ('Temporal Manipulation', 'TM-3', 'Session Manipulation',
     $$Mechanics that artificially extend or constrain individual play sessions, preventing the player from playing at their own pace.$$,
     ARRAY[
        $$Energy refills at 1 unit every 6 minutes — you literally can't play in a single sitting unless you pay to skip the timer.$$,
        $$There's no pause button mid-mission; if you take a phone call you lose the whole run.$$],
     ARRAY[
        $$Online matches obviously can't be paused, but offline mode lets you pause whenever — fair design.$$],
     ARRAY[$$Forced Action$$, $$Obstruction$$],
     ARRAY[$$Zagal: Can't Pause or Save, Wait to Play$$, $$DPGames: Can't Pause or Save, Wait to Play$$]),

    ('Temporal Manipulation', 'TM-4', 'Forced Ad Exposure',
     $$Requiring or heavily incentivizing the player to watch advertisements, consuming their time in exchange for continued play or in-game rewards.$$,
     ARRAY[
        $$Every revive shows a 30-second unskippable ad — dying twice in a row means a minute of advertising before I can play again.$$,
        $$You CAN play without ads, but progress is 5x slower unless you watch a video every match. Not really optional.$$],
     ARRAY[
        $$There's an 'watch ad for a daily bonus' button I can ignore without consequence — completely optional reward.$$],
     ARRAY[$$Nagging$$, $$Disguised Ad$$],
     ARRAY[$$DPGames: Advertisements$$, $$Petrovskaya: Obtrusive Ads$$]),

    ('Predatory Monetization', 'PM-1', 'Pay-to-Progress',
     $$Systems where spending real money provides gameplay advantages or unlocks content/progression that non-paying players cannot access or must invest disproportionate time to reach.$$,
     ARRAY[
        $$Premium players get 2x damage and faster movement — F2P PvP is basically unplayable past silver rank.$$,
        $$You can't even unlock the second city without buying the season pass. Hard paywall after the prologue.$$],
     ARRAY[
        $$There are cosmetic skins you can buy but they don't affect gameplay at all — purely visual stuff.$$],
     ARRAY[$$Pay-to-Play$$, $$Pressured Selling$$],
     ARRAY[$$Zagal: Pay to Win, Pay to Skip$$, $$DPGames: Pay to Win, Pay to Skip, Pay Wall$$, $$Petrovskaya: Pay to Win, Monetization of QoL$$]),

    ('Predatory Monetization', 'PM-2', 'Currency Obfuscation',
     $$Using virtual or multi-layered currency systems to obscure the real-world cost of in-game purchases, making it harder for players to evaluate how much they are actually spending.$$,
     ARRAY[
        $$1 'gem' is supposedly 1 cent but bundles sell at 99/249/499 gems for €0.99/€1.99/€3.99 — try doing the math on what a skin actually costs.$$,
        $$You buy 'platinum', spend platinum on 'crystals', spend crystals on 'tickets'. By the third currency I'd lost track of how much I'd actually spent.$$],
     ARRAY[
        $$Items have clear €/$ prices in the store, no virtual currency middleman — refreshing transparency.$$],
     ARRAY[$$Intermediate Currency$$, $$Hidden Costs$$],
     ARRAY[$$DPGames: Premium Currency$$, $$Petrovskaya: Multi-Layer Currency, Inconvenient Purchase Rates$$]),

    ('Predatory Monetization', 'PM-3', 'Gambling Mechanics',
     $$Systems that use randomized rewards purchased with real money, exploiting the same psychological mechanisms as gambling (variable ratio reinforcement, near-misses, etc.).$$,
     ARRAY[
        $$Drop rate for the SSR character is 0.5%, but the pity timer kicks in at 200 pulls — that's roughly €400 minimum to guarantee one.$$,
        $$Spinning the wheel costs gems; you get whatever it lands on, and 'try again' is one tap away. Pure slot machine.$$],
     ARRAY[
        $$Battle pass rewards are fixed and listed in advance — you know exactly what you get for each tier. No RNG.$$],
     ARRAY[$$Pressured Selling$$],
     ARRAY[$$DPGames: Loot Boxes / Gacha, Gambling / Betting$$, $$Petrovskaya: Loot Boxes / Gambling Mechanics$$]),

    ('Predatory Monetization', 'PM-4', 'Price Manipulation',
     $$Techniques that distort the player's perception of value or cost through anchoring, artificial scarcity, or depreciation of prior purchases.$$,
     ARRAY[
        $$The 'Mega Bundle' is shown crossed-out at €99.99 right next to the €19.99 'special offer' — textbook anchoring.$$,
        $$Shop says 'LIMITED — 3 hours left!' but the same skin came back two weeks later in another 'limited' rotation.$$],
     ARRAY[
        $$Prices are static and shown in your local currency without any 'sale' framing or countdown timers.$$],
     ARRAY[$$False Hierarchy$$, $$Pressured Selling$$, $$Hidden Costs$$, $$Low Stock$$, $$Urgency$$],
     ARRAY[$$DPGames: Anchoring Tricks, Artificial Scarcity, Power Creep$$, $$Petrovskaya: Overpriced Content$$, $$Zagal: Pre-Delivered Content$$]),

    ('Predatory Monetization', 'PM-5', 'Recurring / Compounding Charges',
     $$Monetization systems that create ongoing or escalating financial commitments beyond a single purchase.$$,
     ARRAY[
        $$The battle pass is €10 every season; skip one season and you fall permanently behind the rewards track.$$,
        $$Premium subscription renews monthly and silently upcharged me when the price rose — no email warning.$$],
     ARRAY[
        $$There's an optional one-time DLC purchase, no recurring charges or subscriptions of any kind.$$],
     ARRAY[$$Forced Continuity$$, $$Gamification$$],
     ARRAY[$$DPGames: Recurring Fee$$, $$Petrovskaya: Subscription Advantages, Battle Pass Pressure$$]),

    ('Predatory Monetization', 'PM-6', 'Interface Monetization Traps',
     $$User interface designs that trick, nudge, or facilitate accidental purchases through deceptive placement, confusing flows, or missing confirmation steps.$$,
     ARRAY[
        $$The 'buy now' button is exactly where the 'close' button was two seconds earlier — I bought a 50-gem pack by accident.$$,
        $$Claiming a daily reward triggers a 'CONFIRM SPEND 200 GEMS' dialog with a misleading green button that looks like 'accept reward'.$$],
     ARRAY[
        $$Purchases require a typed PIN and a second confirmation screen — basically impossible to misclick a real-money buy.$$],
     ARRAY[$$Interface Interference$$, $$Trick Questions$$, $$Bad Defaults$$],
     ARRAY[$$DPGames: Accidental Purchases$$, $$Petrovskaya: Malicious Interface Design$$]),

    ('Social Exploitation', 'SE-1', 'Recruitment Pressure',
     $$Mechanics that incentivize or require players to recruit new players from their social network in order to progress, access features, or receive rewards.$$,
     ARRAY[
        $$To unlock the guild boss you need to invite 5 friends who actually start playing — global chat is full of people begging for invites.$$,
        $$The game posts to your friends list automatically saying 'Antonio is recruiting!' without asking for permission.$$],
     ARRAY[
        $$There's a referral code that gives a small one-time bonus if used, but it's tucked away and clearly optional.$$],
     ARRAY[$$Social Pyramid$$, $$Friend Spam$$],
     ARRAY[$$Zagal: Social Pyramid Scheme, Impersonation$$, $$DPGames: Social Pyramid Scheme, Friend Spam / Impersonation$$]),

    ('Social Exploitation', 'SE-2', 'Social Obligation',
     $$Mechanics that create feelings of guilt, duty, or reciprocal debt toward other players, making the player feel they must play or spend to avoid letting others down.$$,
     ARRAY[
        $$Your guild loses points if you don't log in daily, so the other 49 members guilt you into playing even when you don't want to.$$,
        $$If your co-op partner doesn't show up you BOTH lose three days of progress — pure peer pressure to keep logging in.$$],
     ARRAY[
        $$Co-op missions reward both players when you play together, but you're not penalised at all for soloing.$$],
     ARRAY[$$Shaming$$, $$Social Proof$$, $$Forced Action$$],
     ARRAY[$$Zagal: Social Obligation$$, $$DPGames: Social Obligation / Guilt, Reciprocity$$]),

    ('Social Exploitation', 'SE-3', 'Competitive Pressure',
     $$Systems that exploit players' competitive nature by creating environments where spending money is the primary way to compete, or by using matchmaking to showcase the advantages of paying.$$,
     ARRAY[
        $$Free players keep getting matched against premium subscribers and losing — the message is clear: pay or lose.$$,
        $$The leaderboard is locked behind a €30 'competitive pack' that gives you a 15% stat boost over anyone who didn't buy it.$$],
     ARRAY[
        $$Ranked PvP uses MMR-based matchmaking; paying doesn't influence skill brackets at all.$$],
     ARRAY[$$Pressured Selling$$, $$Social Proof$$, $$Shaming$$],
     ARRAY[$$Zagal: Monetized Rivalries$$, $$DPGames: Competition$$, $$Petrovskaya: Matchmaking Against Paying Players$$]),

    ('Psychological Exploitation', 'PE-1', 'Loss Aversion Exploitation',
     $$Mechanics that leverage the player's fear of losing progress, items, or opportunities — exploiting the psychological principle that losses feel more painful than equivalent gains.$$,
     ARRAY[
        $$I've spent €200 over six months — quitting now feels like throwing all that money away, which is exactly what they want.$$,
        $$Limited-time event ends in 6 hours, items will NEVER return — anxiety-driven purchase, classic FOMO.$$],
     ARRAY[
        $$All cosmetics rotate back into the shop every couple of months, so missing one event isn't a big deal.$$],
     ARRAY[$$Roach Motel$$, $$Urgency$$, $$Scarcity and Popularity Claims$$],
     ARRAY[$$DPGames: Invested / Sunk Cost, Fear of Missing Out (FOMO)$$, $$Petrovskaya: FOMO / Limited-Time Offers$$]),

    ('Psychological Exploitation', 'PE-2', 'Completionism Exploitation',
     $$Mechanics that exploit the player's desire to collect, complete, or achieve everything, creating artificial compulsion to engage beyond enjoyment.$$,
     ARRAY[
        $$The 100% achievement requires collecting 500 of a currency that drops at 1% rate — designed entirely for completionists.$$,
        $$The collection tab shows 87/100 done, dangling that last 13% specifically to keep me grinding well past when I'd stopped enjoying it.$$],
     ARRAY[
        $$There are optional collectibles but the game makes it clear they're not required and won't affect the story.$$],
     ARRAY[$$Gamification$$],
     ARRAY[$$DPGames: Badges / Completionism$$]),

    ('Psychological Exploitation', 'PE-3', 'Cognitive Bias Exploitation',
     $$Mechanics that exploit specific cognitive biases to make players misjudge probabilities, costs, or their own agency.$$,
     ARRAY[
        $$You 'pick' a chest from three identical boxes — but the reward was determined before you tapped. The choice is fake, just illusion of control.$$,
        $$Drop rate is shown as '~15%' but in practice it feels closer to 3% — they exploit optimism bias by showcasing the rare drops prominently.$$],
     ARRAY[
        $$Drop rates are published transparently and have been verified by community datamining.$$],
     ARRAY[$$Feedforward Ambiguity$$, $$(De)contextualizing Cues$$, $$Pressured Selling$$],
     ARRAY[$$DPGames: Illusion of Control, Optimism Bias Exploitation, Variable Rewards$$]),

    ('Psychological Exploitation', 'PE-4', 'Sensory Manipulation',
     $$Using visual effects, sound design, animations, or other sensory elements to trigger emotional responses that override rational decision-making about spending or engagement.$$,
     ARRAY[
        $$Opening a loot box has these glittery near-miss animations and a triumphant fanfare even for trash drops — engineered to feel like a win.$$,
        $$Every purchase triggers a 5-second celebration screen with confetti and rising chords, like a slot machine win.$$],
     ARRAY[
        $$Purchases just show a quiet 'item added to inventory' notification with no fanfare or animation.$$],
     ARRAY[$$Emotional or Sensory Manipulation$$],
     ARRAY[$$DPGames: Aesthetic Manipulation$$]),

    ('Deceptive Representation', 'DR-1', 'Misleading Marketing',
     $$Advertising, store listings, or promotional materials that misrepresent the actual gameplay, graphics, content, or experience the player will have.$$,
     ARRAY[
        $$The Facebook ads show a totally different game — actual gameplay is nothing like the promised tactical battles.$$,
        $$Trailer features mechanics and characters that aren't even in the released game. Pure bait.$$],
     ARRAY[
        $$Marketing accurately represents the gameplay; what you see in trailers is what you get.$$],
     ARRAY[$$Bait and Switch$$],
     ARRAY[$$Petrovskaya: Misleading Advertising, Unrealistic Ad Presentation$$]),

    ('Deceptive Representation', 'DR-2', 'Post-Purchase Deception',
     $$Changes or discrepancies between what was promised/previewed at the point of purchase and what the player actually receives or experiences after spending.$$,
     ARRAY[
        $$I paid €25 for the meta hero and they nerfed it to the ground two weeks later — never refunded the difference.$$,
        $$The cosmetic looks completely different in-game than in the store preview — much lower quality and the colours are off.$$],
     ARRAY[
        $$When they balance-changed a character they offered a free refund window — fair treatment of paying players.$$],
     ARRAY[$$Bait and Switch$$, $$Hiding Information$$, $$Hidden Costs$$],
     ARRAY[$$Petrovskaya: Nerfing After Purchase, Low-Quality Cosmetics$$, $$Zagal: Pre-Delivered Content$$])
) AS v(family, code, name, description, examples, counter_examples, gray_mapping, source_mapping)
JOIN taxonomy_high_levels p ON p.name = v.family
ON CONFLICT (code, version) DO NOTHING;
