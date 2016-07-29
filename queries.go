package main 


const(
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
  
)