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

	INSERT_WORDS = `insert into words values (DEFAULT, "Achiever", "Determined", "They are determined to do whatever it takes to achieve results."),
  (DEFAULT, "Achiever", "Self-motivated", "They are internally driven to get things done."),
  (DEFAULT, "Activator", "Action-oriented", "They would rather take action than continue to debate and discuss."),
  (DEFAULT, "Activator", "Starter", "As soon as a decision was made, they want to start turning thoughts into action."),
  (DEFAULT, "Adaptability", "Easygoing", "They enjoyed going with the flow and found it easy to adapt to new circumstances"),
  (DEFAULT, "Adaptability", "Spontaneous", "They were able to spontaneously adapt to new information and opportunities "),
  (DEFAULT, "Analytical", "Logical", "They examine factors and patterns to understand the truth and its implications."),
  (DEFAULT, "Analytical", "Researcher", "They researched to uncover essential facts necessary to attain excellence."),
  (DEFAULT, "Arranger", "Organizer", "They were really good at putting together the right team of diverse people to address a complex problem."),
  (DEFAULT, "Arranger", "Simplifier", "A master arranger, they enjoy taking a complex situation and making it smoother.  "),
  (DEFAULT, "Belief", "Ethical", "They draw strength from their values and convictions in doing what is right. "),
  (DEFAULT, "Belief", "Purposeful", "Their motivaction comes from deeply held ideals and sense of purpose."),
  (DEFAULT, "Command", "Decisive", "They are willing to speak up and make a decision."),
  (DEFAULT, "Command", "Directive", "They are able to take charge and bring clarity to a situation"),
  (DEFAULT, "Communicator", "Conversationalist", "They find captivating words for their thoughts and feelings that are relatable to others. "),
  (DEFAULT, "Communicator", "Storyteller", "They tell stories and present in a way that brings ideas and events to life."),
  (DEFAULT, "Competition", "Aspirational", "They aspire to be the top performer and create a culture of winning. "),
  (DEFAULT, "Competition", "Competitive", "They enjoy contests, measure their progress against the performance of others, and strive to win."),
  (DEFAULT, "Connectedness", "Bridge builder", "They build bridges between people and groups, showing them how to relate to, and rely on each other."),
  (DEFAULT, "Connectedness", "Tolerant", "They are considerate, caring, and accepting of everyone."),
  (DEFAULT, "Consistency", "Fair", "They use policies, procedures and rules to ensure everyone is treated fairly."),
  (DEFAULT, "Consistency", "Predictable", "They prefer predictable, repeatable, reliable processes that result in equitable outcomes."),
  (DEFAULT, "Context", "Historian", "They are more confident when they understand the history of a given situation."),
  (DEFAULT, "Context", "Retrospective", "They value hindsight and use past learnings as blueprints for future direction."),
  (DEFAULT, "Deliberative", "Careful", "When making a decision, they take time to identify and work out all potential risks."),
  (DEFAULT, "Deliberative", "Conscientious", "Mindful of the dangers of a bad decision, they thoroughly examine every alternative to ensure they make the right choice."),
  (DEFAULT, "Developer", "Mentor", "They are great at helping people learn and grow into their potential."),
  (DEFAULT, "Developer", "Encouraging", "They are encouraging and know how to bring out the best in others."),
  (DEFAULT, "Discipline", "Organized", "They bring a high level of stability and order to their work by focusing on timelines and deadlines."),
  (DEFAULT, "Discipline", "Structured", "They thrive in an orderly environment with routine, structure, and precision.  "),
  (DEFAULT, "Empathy", "Perceptive", "Their ability to perceive the feelings of others allows them to provide comfort and stability."),
  (DEFAULT, "Empathy", "Understanding", "They can sense and understand other people's feelings and give voice to multiple perspectives."),
  (DEFAULT, "Focus", "Efficient", "Their efficient work style helps keep the entire team focused and on track. "),
  (DEFAULT, "Focus", "Goal-oriented", "They take aim on a goal and focus their effort for efficiency. "),
  (DEFAULT, "Futuristic", "Imaginative", "Their ability to imagine and describe a better future energizes the people around them"),
  (DEFAULT, "Futuristic", "Visionary", "They can imagine what is possible in the future and inspire others with the picture they paint."),
  (DEFAULT, "Harmony", "Agreeable", "They prefer harmony, see commonalities and steer others toward agreement and away from conflict. "),
  (DEFAULT, "Harmony", "Peacekeeper", "They value peace and naturally steer others away from confrontation. "),
  (DEFAULT, "Ideation", "Free thinker", "They enjoy coming up with new ideas "),
  (DEFAULT, "Ideation", "Innovative", "Their innovative approach to problems and projects can be a spontaneous source of new and valuable ideas."),
  (DEFAULT, "Includer", "Accepting", "They accept and appreciate other people and want to bring them into the group."),
  (DEFAULT, "Includer", "Engaging", "They thoughtfully engage others to increase their participation while also up-leveling acceptance of diversity."),
  (DEFAULT, "Individualization", "Appreciative", "They notice and appreciate the unique characteristics of each person and customized their approach accordingly. "),
  (DEFAULT, "Individualization", "Relatable", "They see each person’s unique style and motivation and draw out the best in each by relating to them as individuals."),
  (DEFAULT, "Input", "Collector", "They collect certain things, such as ideas, books, memorabilia, quotations, or facts because they find them interesting. "),
  (DEFAULT, "Input", "Resourceful", "They are curious, gather information, and store knowledge that can be helpful. "),
  (DEFAULT, "Intellection", "Deep Thinker", "They like to exercise their brain, stretching in multiple directions."),
  (DEFAULT, "Intellection", "Introspective", "They take time to ponder and process, developing wisdom and clarity as a result."),
  (DEFAULT, "Learner", "Curious", "They are curious and love to continuously learn and improve."),
  (DEFAULT, "Learner", "Studious", "They are energized by the process of studying and learning something new. "),
  (DEFAULT, "Maximizer", "Coach", "They love to help others become excited about their potential and focus on areas of strengths."),
  (DEFAULT, "Maximizer", "Strengths-oriented", "They focus on strengths as a way to help people and groups live up to their potential."),
  (DEFAULT, "Positivity", "Enthusiastic", "They make everything more exciting and lighten the spirits of those around them."),
  (DEFAULT, "Positivity", "Optimistic", "They are generous with praise, quick to smile, and always on the lookout for the upside of a situation."),
  (DEFAULT, "Relator", "Friendly", "They know and relate to all kinds of people and have close relationships with a smaller group."),
  (DEFAULT, "Relator", "Personal", "They form solid, genuine, and mutually rewarding relationships that are close, caring, and trusting. "),
  (DEFAULT, "Responsibility", "Dependable", "They keep their promises, honor their commitments, and work hard to fulfill all of their responsibilities."),
  (DEFAULT, "Responsibility", "Trustworthy", "They are people of their word, and others know they can rely on and trust them."),
  (DEFAULT, "Restorative", "Problem solver", "They are good at figuring out what is wrong and resolving it."),
  (DEFAULT, "Restorative", "Troubleshooter", "They enjoy the challenge of analyzing symptoms, identifying what is wrong, and finding a solution."),
  (DEFAULT, "Self-assurance", "Confident", "They have faith in their strengths and abilities and trust their own instincts in forging ahead, even on risky paths. "),
  (DEFAULT, "Self-assurance", "Fearless", "They have an inner sense of certainty that affirms their direction and decisions and instills confidence in others, even in the midst of turbulence."),
  (DEFAULT, "Significance", "Recognized", "They are determined to make important contributions worthy of recognition."),
  (DEFAULT, "Significance", "Visible", "They enjoy being visible and known for their unique talents."),
  (DEFAULT, "Strategic", "Navigator", "They chart all the possible paths and navigate the best way to achieve the strategic goals."),
  (DEFAULT, "Strategic", "Wayfinder", "They are able to see the big picture, consider multiple approaches, and find the best way forward."),
  (DEFAULT, "Woo", "Networker", "They enjoy being around people, starting up conversations, and getting acquainted."),
  (DEFAULT, "Woo", "Social", "They enjoy getting to know someone new.")`
)
