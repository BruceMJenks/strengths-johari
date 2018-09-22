package main

const (
	// Adjectives that are selected by both the participant and his or her peers are placed into the Open or Arena quadrant.
	JOHARI_ARENA_QUERY = `SELECT w.word FROM words w JOIN (
    SELECT DISTINCT p.word FROM peers p JOIN subjects s ON s.session = p.session WHERE p.session = '%s' AND p.word = s.word ) sub
    ON sub.word = w.wid order by 1`

	// Adjectives that are not selected by subjects but only by their peers are placed into the Blind Spot quadrant
	JOHARI_BLIND_QUERY = `SELECT DISTINCT w.word FROM peers p JOIN words w ON w.wid = p.word WHERE p.session = '%s' AND p.word not in 
                  ( SELECT s.word FROM subjects s JOIN peers p on s.session = p.session WHERE s.session = '%s') order by 1`

	// Adjectives selected only by subjects, but not by any of their peers, are placed into the Hidden or Façade quadrant
	JOHARI_FACADE_QUERY = `SELECT DISTINCT w.word FROM subjects s JOIN words w on w.wid = s.word WHERE s.session = '%s' AND s.word not in 
                  ( SELECT p.word from peers p JOIN subjects s ON s.session = p.session WHERE p.session = '%s') order by 1
                  `

	// Adjectives that were not selected by either subjects or their peers remain in the Unknown quadrant
	JOHARI_UNKOWN_QUERY = `SELECT DISTINCT word FROM words WHERE wid NOT IN 
                  ( SELECT p.word FROM peers p WHERE p.session = '%s'
                    UNION 
                    SELECT s.word FROM subjects s WHERE s.session = '%s')`

	CLIFTON_ARENA_QUERY = `
    SELECT DISTINCT w.theme FROM words w JOIN (
      SELECT DISTINCT p.word FROM peers p JOIN subjects s ON s.session = p.session WHERE p.session = '%s' AND p.word = s.word ) sub
      ON sub.word = w.wid order by 1
  `
	CLIFTON_BLIND_QUERY = `
    SELECT DISTINCT w.theme FROM peers p JOIN words w ON w.wid = p.word WHERE p.session = '%s' AND p.word not in 
        ( SELECT s.word FROM subjects s JOIN peers p on s.session = p.session WHERE s.session = '%s') order by 1
  `
	CLIFTON_FACADE_QUERY = `
    SELECT DISTINCT w.theme FROM subjects s JOIN words w on w.wid = s.word WHERE s.session = '%s' AND s.word not in 
      ( SELECT p.word from peers p JOIN subjects s ON s.session = p.session WHERE p.session = '%s') order by 1
  `

	CLIFTON_UNKOWN_QUERY = `
  SELECT DISTINCT theme FROM words WHERE theme NOT IN 
                  ( SELECT DISTINCT w.theme FROM words w JOIN peers p ON p.word = w.wid WHERE p.session = '%s'
                    UNION 
                    SELECT w.theme FROM words w JOIN subjects s ON s.word = w.wid WHERE s.session = '%s')
  `

	PREVIOUS_WINDOWS_QUERY = `SELECT DISTINCT s.session, s.nickname FROM sessions s JOIN subjects sj ON sj.session = s.session WHERE uid = %d ORDER by s.timecreated`

	CREATE_USERS_TABLE = `CREATE TABLE IF NOT EXISTS users (
    id serial,
    username varchar(255),
    refreshtoken varchar(4096)
  )`

	CREATE_WORDS_TABLE = `CREATE TABLE IF NOT EXISTS words (
    wid serial,
    theme text,
    word text,
    description text )`

	SELECT_WORDS_TABLE = `SELECT count(*) FROM words`

	CREATE_SUBJECTS_TABLE = `CREATE TABLE IF NOT EXISTS subjects (
    uid int,
    session text,
    word int
  )`

	CREATE_PEERS_TABLE = `CREATE TABLE IF NOT EXISTS peers (
    uid int, 
    session text, 
    word int
  )`

	CREAT_SESSIONS_TABLE = `CREATE TABLE IF NOT EXISTS sessions (
    timecreated timestamp,
    session text,
    nickname text
  )`

	DROP_WORDS_TABLE    = `DROP TABLE IF EXISTS words`
	DROP_PEERS_TABLE    = `DROP TABLE IF EXISTS peers`
	DROP_USERS_TABLE    = `DROP TABLE IF EXISTS users`
	DROP_SUBJECTS_TABLE = `DROP TABLE IF EXISTS subjects`
	DROP_SESSIONS_TABLE = `DROP TABLE IF EXISTS sessions`

	SELECT_USER_QUERY          = `SELECT * from users where username = '%s' LIMIT 1`
	SELECT_USERNAME_QUERY      = `SELECT username FROM users where username = '%s' LIMIT 1`
	SELECT_USERID_QUERY        = `SELECT id FROM users WHERE username = '%s' LIMIT 1`
	SELECT_USER_PASSWORD_QUERY = `SELECT refreshtoken FROM users where username = '%s' limit 1`
	INSERT_USER_QUERY          = `INSERT INTO users VALUES (DEFAULT, ?, ?)`

	SELECT_SUBJECT_SESSION_QUERY = `SELECT session FROM subjects WHERE uid = %d and session = '%s' LIMIT 1`

	SELECT_WORDLIST_QUERY = `SELECT word, description FROM words order by 1`

	INSERT_WORDS = `insert into words values (DEFAULT, "Achiever", "Productive", "She was not satisfied with just keeping busy. She was keen on creating productive assets."),
  (DEFAULT, "Activator", "Catalytic", "Just as an enzyme lowers the energy of the transition state, so as soon as he walked in the room we were off and running."),
  (DEFAULT, "Adaptability", "Spontaneous", "I rarely saw her get perturbed when plans went awry. Rather she would effortlessly roll with the punches."),
  (DEFAULT, "Analytical", "Objective", "His searching out of the truth was so impartial that at times it seemed impersonal."),
  (DEFAULT, "Arranger", "Resourceful", "She was able to deal promptly and skillfully with almost any situation."),
  (DEFAULT, "Belief", "Certain", "Hers was a life of confidence. She was free of doubt and reservation."),
  (DEFAULT, "Command", "Decisive", "You could count on him to put an end to controversy. He did so without hesitation."),
  (DEFAULT, "Comunication", "Expressive", "He had an uncanny knack for story telling. He could take the most convoluted situation and turn into something we all got."),
  (DEFAULT, "Competition", "Scorekeeping", "She kept track of the achievements of those around her. Not primarily to be the best but to urge herself to expend even higher levels of exertion."),
  (DEFAULT, "Connectedness", "Sensemaking", "After the reorg, we stalled. Everybody had an opinion. He helped us tie it all together. After that, we remembered who we were and we got to work again."),
  (DEFAULT, "Consistency", "Just", "In dealing with others, he was guided by principle. There was not a crooked bone in him."),
  (DEFAULT, "Context", "Historical", "You could count on him to bring the facts, documented and attested to."),
  (DEFAULT, "Deliberative", "Serious", "I know that when I bring her a worry, she will not treat it lightly. Rather, she will treat it as a weighty matter."),
  (DEFAULT, "Developer", "Investing", "Like a venture capitalist, she was a shrewd investor. Only she didn’t invest in companies. She invested in people."),
  (DEFAULT, "Discipline", "Orderly", "His life was governed by strict attention to method. Even his sock drawer was a study in tidiness."),
  (DEFAULT, "Empathy", "Aware", "She was so cognizant of the feelings of others, we often asked her to read the room for us."),
  (DEFAULT, "Focus", "Selective", "He was careful about what he committed himself to, knowing that when he did, he would give it his all."),
  (DEFAULT, "Futuristic", "Broad", "His horizons were so expansive that they seemed without limit."),
  (DEFAULT, "Harmony", "Collaborative", "When I asked him how he preferred to work, he replied, “Together.”"),
  (DEFAULT, "Ideation", "Artistic", "Her ideas were so imaginative that, for her, there was no box."),
  (DEFAULT, "Includer", "Welcoming", "She was our best new hire ambassador, gladly receiving new arrivals without obligation."),
  (DEFAULT, "Individualization", "Astute", "Her discernment was remarkable. She could see things in people that others would overlook."),
  (DEFAULT, "Input", "Inquisitive", "She was so thirsty for knowledge; she never seemed to run out of questions to ask."),
  (DEFAULT, "Intellection", "Reflective", "He relished the times where he could quietly bend his thoughts backward, as looking in a mirror. Not to solve some problem per se, but just to gaze."),
  (DEFAULT, "Learner", "Curious", "There seemed to be no “off” button for her inquiring mind. The more she learned, the more she became aware of what she didn’t know."),
  (DEFAULT, "Maximizer", "Discriminating", "He was an excellent judge of human potential. He had a keen eye for promise."),
  (DEFAULT, "Positivity", "Optimistic", "Her expectation that events would turn out favorably was so infectious, that it was hard not to smile and go along."),
  (DEFAULT, "Relator", "Genuine", "She got on well with others because her words and actions always aligned. In her, there was no pretense."),
  (DEFAULT, "Responsibility", "Diligent", "When she says, “I got this.” She means she will stick with it until it is done."),
  (DEFAULT, "Restorative", "Fixer", "In the movies, people call in the Fixer when matters are miserably broken and out of control. She is our Fixer, except her methods are always above board. She makes it right."),
  (DEFAULT, "Self-assurance", "Stable", "He is steadfast, constant, and unwavering. We have come to rely upon him to be even-keeled no matter what."),
  (DEFAULT, "Significance", "Successful", "She seems to strongly motivated by visible markers of achievement and progress. I made sure we included these in her development plan."),
  (DEFAULT, "Strategic", "Anticipating", "As he looked out over the sea of facts, he was able to take hold of what would likely happen with relative ease."),
  (DEFAULT, "Woo", "Charming", "His manner was so engaging that soon all of us were hanging on his every word."),
  (DEFAULT, "Achiever", "Self-motivated", "She did not need others to prod her; she was keen to complete the task at hand."),
  (DEFAULT, "Activator", "Impatient", "She was restless; always eager to “get to getting.”"),
  (DEFAULT, "Adaptability", "Flexible", "Far from being rigid and brittle, he was able to morph himself into whatever shape the situation called for."),
  (DEFAULT, "Analytical", "Questioning", "Her intellectual curiosity was relentless. She would not rest until the proffered explanation was either verified or set aside."),
  (DEFAULT, "Arranger", "Configuring", "As events emerged, he was able to weave them together into a pattern that made sense."),
  (DEFAULT, "Belief", "Principled", "His conduct seemed to follow a truth that withstood the test of time."),
  (DEFAULT, "Command", "Assertive", "He did not speak often, but when he did, his confidence was obvious."),
  (DEFAULT, "Comunication", "Captivating", "In whatever she said, wrote, or presented, she always took hold of our attention."),
  (DEFAULT, "Competition", "Driven", "No contest was too small for him; he was compelled to win."),
  (DEFAULT, "Connectedness", "Perceptive", "We came to rely on her keen insights to tie up the loose ends of our story."),
  (DEFAULT, "Consistency", "Fair", "He did not seem to harbor a trace of bias. He was a man free of guile."),
  (DEFAULT, "Context", "Grounded", "For her, the backstory provided a firm foundation for going forward."),
  (DEFAULT, "Deliberative", "Vigilant", "Ever watchful for danger, she remained on alert for risks both large and small."),
  (DEFAULT, "Developer", "Observant", "He was a watchful person when it came to developing talents in other people."),
  (DEFAULT, "Discipline", "Meticulous", "Her life was an orderly one. She approached each task with the same precision and care."),
  (DEFAULT, "Empathy", "Intuitive", "In dealing with others, he seemed to have a direct perception of their truth."),
  (DEFAULT, "Focus", "Persevering", "Once he accepted a task, we could rely upon him to stick with it in spite of difficulty or discouragement."),
  (DEFAULT, "Futuristic", "Vivid", "We could not stop listening as he painted a picture of what could be – it was so bright and full of life."),
  (DEFAULT, "Harmony", "Efficient", "He was a true team player. Somehow he seemed to guide others to nearly frictionless solutions."),
  (DEFAULT, "Ideation", "Innovative", "Whether she was making something anew or renewing something old, for her, change was a good thing."),
  (DEFAULT, "Includer", "Accepting", "She was known for freely taking others in."),
  (DEFAULT, "Individualizaton", "Diverse", "She was attuned to even the smallest differences between others."),
  (DEFAULT, "Input", "Knowledgeable", "If it was knowable, he seemed to know it."),
  (DEFAULT, "Intellection", "Solitary", "His rich interior world drove him to seek periods of quiet reflection as much as it did to seek company."),
  (DEFAULT, "Learner", "Competent", "Her knowledge made her more than capable to meet their need."),
  (DEFAULT, "Maximizer", "Builder", "He was known for taking plans and prototypes and turning them into enduring structures."),
  (DEFAULT, "Positivity", "Energetic", "She was a dynamo. Others could not help but be inspired by her can-do attitude."),
  (DEFAULT, "Relator", "Authentic", "He was the genuine article. What you saw was what you got."),
  (DEFAULT, "Responsibility", "Dutiful", "She performed even the most mundane tasks with care and cheerfulness."),
  (DEFAULT, "Restorative", "Responsive", "He responded to issues in kind, with a ready and sympathetic ear."),
  (DEFAULT, "Self-assurance", "Intense", "At times, his speech was so impassioned that he could appear as a zealot."),
  (DEFAULT, "Significance", "Influential", "Consequence was his watchword. Everything he put his hand to must matter."),
  (DEFAULT, "Strategic", "Clear", "When we listened to her explanations, nothing was obscure or darkened."),
  (DEFAULT, "Woo", "Gregarious", "He was companionable and sociable. No matter what group he visited, he always seemed to fit in.")`
)
