-- Add CSP (Content Server Protocol) assist/input fields to laps
ALTER TABLE laps ADD COLUMN abs_level INTEGER;
ALTER TABLE laps ADD COLUMN tc_level INTEGER;
ALTER TABLE laps ADD COLUMN stability_control REAL;
ALTER TABLE laps ADD COLUMN auto_shifting INTEGER;
ALTER TABLE laps ADD COLUMN input_method INTEGER;
ALTER TABLE laps ADD COLUMN tyre_compound INTEGER;
