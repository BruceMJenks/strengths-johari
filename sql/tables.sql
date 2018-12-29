CREATE TABLE users (
	id serial,
	username varchar(255),
	refreshtoken varchar(4096)
);

CREATE TABLE words (
	wid serial,
	theme text,
	word text
);

CREATE TABLE subjects (
	uid int,
	session text,
	word int
);

CREATE TABLE peers (
	uid int, 
	session text, 
	word int
);

CREATE TABLE sessions (
	timecreated timestamp,
	session text,
	nickname text
);


insert into words values (DEFAULT, "Achiever", "Productive"),
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
(DEFAULT, "Woo", "Gregarious");