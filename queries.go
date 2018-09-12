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

	PREVIOUS_WINDOWS_QUERY = `"SELECT DISTINCT s.session, s.nickname FROM sessions s JOIN subjects sj ON sj.session = s.session WHERE uid = %d ORDER by s.timecreated"`

	CREATE_USERS_TABLE = `CREATE TABLE users (
    id serial,
    username varchar(255),
    refreshtoken varchar(4096)
  )`

	CREATE_WORDS_TABLE = `CREATE TABLE words (
    wid serial,
    theme text,
    word text
  )`

	CREATE_SUBJECTS_TABLE = `CREATE TABLE subjects (
    uid int,
    session text,
    word int
  )`

	CREATE_PEERS_TABLE = `CREATE TABLE peers (
    uid int, 
    session text, 
    word int
  )`

	INSERT_WORDS = `insert into words values (DEFAULT, "Achiever", "Productive"),
  (DEFAULT, "Activator", "Catalytic"),
  (DEFAULT, "Adaptability", "Spontaneous"),
  (DEFAULT, "Analytical", "Objective"),
  (DEFAULT, "Arranger", "Resourceful"),
  (DEFAULT, "Belief", "Certain"),
  (DEFAULT, "Command", "Decisive"),
  (DEFAULT, "Comunication", "Expressive"),
  (DEFAULT, "Competition", "Scorekeeping"),
  (DEFAULT, "Connectedness", "Sensemaking"),
  (DEFAULT, "Consistency", "Just"),
  (DEFAULT, "Context", "Historical"),
  (DEFAULT, "Deliberative", "Serious"),
  (DEFAULT, "Developer", "Investing"),
  (DEFAULT, "Discipline", "Orderly"),
  (DEFAULT, "Empathy", "Aware"),
  (DEFAULT, "Focus", "Selective"),
  (DEFAULT, "Futuristic", "Broad"),
  (DEFAULT, "Harmony", "Collaborative"),
  (DEFAULT, "Ideation", "Artistic"),
  (DEFAULT, "Includer", "Welcoming"),
  (DEFAULT, "Individualizaton", "Astute"),
  (DEFAULT, "Input", "Inquisitive"),
  (DEFAULT, "Intellection", "Reflective"),
  (DEFAULT, "Learner", "Curious"),
  (DEFAULT, "Maximizer", "Discriminating"),
  (DEFAULT, "Positivity", "Optimistic"),
  (DEFAULT, "Relator", "Genuine"),
  (DEFAULT, "Responsibility", "Diligent"),
  (DEFAULT, "Restorative", "Fixer"),
  (DEFAULT, "Self-assurance", "Stable"),
  (DEFAULT, "Significance", "Successful"),
  (DEFAULT, "Strategic", "Anticipating"),
  (DEFAULT, "Woo", "Charming"),
  (DEFAULT, "Achiever", "Self-motivated"),
  (DEFAULT, "Activator", "Impatient"),
  (DEFAULT, "Adaptability", "Flexible"),
  (DEFAULT, "Analytical", "Questioning"),
  (DEFAULT, "Arranger", "Configuring"),
  (DEFAULT, "Belief", "Principled"),
  (DEFAULT, "Command", "Assertive"),
  (DEFAULT, "Comunication", "Captivating"),
  (DEFAULT, "Competition", "Driven"),
  (DEFAULT, "Connectedness", "Perceptive"),
  (DEFAULT, "Consistency", "Fair"),
  (DEFAULT, "Context", "Grounded"),
  (DEFAULT, "Deliberative", "Vigilant"),
  (DEFAULT, "Developer", "Observant"),
  (DEFAULT, "Discipline", "Meticulous"),
  (DEFAULT, "Empathy", "Intuitive"),
  (DEFAULT, "Focus", "Persevering"),
  (DEFAULT, "Futuristic", "Vivid"),
  (DEFAULT, "Harmony", "Efficient"),
  (DEFAULT, "Ideation", "Innovative"),
  (DEFAULT, "Includer", "Accepting"),
  (DEFAULT, "Individualizaton", "Diverse"),
  (DEFAULT, "Input", "Knowledgeable"),
  (DEFAULT, "Intellection", "Solitary"),
  (DEFAULT, "Learner", "Competent"),
  (DEFAULT, "Maximizer", "Builder"),
  (DEFAULT, "Positivity", "Energetic"),
  (DEFAULT, "Relator", "Authentic"),
  (DEFAULT, "Responsibility", "Dutiful"),
  (DEFAULT, "Restorative", "Responsive"),
  (DEFAULT, "Self-assurance", "Intense"),
  (DEFAULT, "Significance", "Influential"),
  (DEFAULT, "Strategic", "Clear"),
  (DEFAULT, "Woo", "Gregarious")`
)
