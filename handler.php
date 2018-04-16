<?php

namespace CalProxy;

use ICal\Event;

class handler {

    /**
     * Parse the event and do the replacement and optimizations
     * @param $e Event a single ical event that should be cleaned up
     */
    public static function cleanEvent(Event &$e) {
        $event = new \Eluceo\iCal\Component\Event();

        //Strip added slashes by the parser
        $summary = stripcslashes($e->summary);
        $description = stripcslashes($e->description);
        $location = stripcslashes($e->location);

        //Remember the old title in the description
        $event->setDescription($summary . "\n" . $description);
        $event->setLocation($location);

        //Remove the TAG and anything after e.g.: (IN0001) or [MA0001]
        $summary = preg_replace('/([\(\[](?:(?:IN|MA|WI)\d+,?\s?)+[\)\]]).+/', '', $summary);

        //Some common replacements: yes its a long list
        $searchReplace = [];
        $searchReplace['Tutorübungen'] = 'TÜ';
        $searchReplace['Grundlagen'] = 'G';
        $searchReplace['Datenbanken'] = 'DB';
        $searchReplace['Zentralübungen'] = 'ZÜ';
        $searchReplace['Zentralübung'] = 'ZÜ';
        $searchReplace['Vertiefungsübungen'] = 'VÜ';
        $searchReplace['Übungen'] = 'Ü';
        $searchReplace['Übung'] = 'Ü';
        $searchReplace['Exercises'] = 'EX';
        $searchReplace['Planen und Entscheiden in betrieblichen Informationssystemen - Wirtschaftsinformatik 4'] = 'PLEBIS';
        $searchReplace['Planen und Entscheiden in betrieblichen Informationssystemen'] = 'PLEBIS';
        $searchReplace['Statistics for Business Administration (with Introduction to  R)'] = 'Stats';
        $searchReplace['Kostenrechnung für Wirtschaftsinformatik und Nebenfach'] = 'KR';
        $searchReplace['Kostenrechnung'] = 'KR';
        $searchReplace['Mathematische Behandlung der Natur- und Wirtschaftswissenschaften (Mathematik 1)'] = 'MBNW';
        $searchReplace['Einführung in die Wirtschaftsinformatik'] = 'WINFO';
        $searchReplace['Betriebssysteme und Systemsoftware'] = 'BS';
        $searchReplace['Einführung in die Informatik '] = 'INFO';
        $searchReplace['Praktikum: Grundlagen der Programmierung'] = 'PGP';
        $searchReplace['Einführung in die Rechnerarchitektur'] = 'ERA';
        $searchReplace['Einführung in die Softwaretechnik'] = 'EIST';
        $searchReplace['Algorithmen und Datenstrukturen'] = 'AD';
        $searchReplace['Rechnernetze und Verteilte Systeme'] = 'RNVS';
        $searchReplace['Einführung in die Theoretische Informatik'] = 'THEO';
        $searchReplace['Diskrete Strukturen'] = 'DS';
        $searchReplace['Diskrete Wahrscheinlichkeitstheorie'] = 'DWT';
        $searchReplace['Numerisches Programmieren'] = 'NumProg';
        $searchReplace['Lineare Algebra für Informatik'] = 'LinAlg';
        $searchReplace['Analysis für Informatik'] = 'Analysis';
        $searchReplace[' der Künstlichen Intelligenz'] = 'KI';
        $searchReplace['Advanced Topics of Software Engineering'] = 'ASE';
        $searchReplace['Praktikum - iPraktikum, iOS Praktikum'] = 'iPraktikum';

        //Do the replacement
        $summary = strtr($summary, $searchReplace);

        //Remove some stuff which is not really needed
        $summary = str_replace(['Standardgruppe', 'PR, ', 'VO, ', 'FA, ', 'VI, ', 'TT, ', 'UE, '], '', $summary);

        //Try to make sense out of the location
        if (!empty($location)) {
            preg_match('/^(.*?,)/', $location, $matches);
            if (preg_match('/56\d{2}\.((EG)|\d{2})\.\d+/', $location)===1) {
                // Informatik
                self::switchLocation($event, $location, $matches[1].' Boltzmannstraße 3, 85748 Garching bei München');
            } else if (preg_match('/55\d{2}\.((EG)|\d{2})\.\d+/', $location)===1) {
                // Maschbau
                self::switchLocation($event, $location, $matches[1].' Boltzmannstraße 15, 85748 Garching bei München');
            } else if (preg_match('/8101\.((EG)|\d{2})\.\d+/', $location)===1) {
                // Hochbrück - Physics
                self::switchLocation($event, $location, $matches[1].' Parkring 11-13, 85748 Garching bei München');
            } else if (preg_match('/8102\.((EG)|\d{2})\.\d+/', $location)===1) {
                // Hochbrück - Informatik
                self::switchLocation($event, $location, $matches[1].' Parkring 35-39, 85748 Garching bei München');
            } else if (preg_match('/51\d{2}\.((EG)|\d{2})\.\d+/', $location)===1) {
                // Physik
                self::switchLocation($event, $location, $matches[1].' James-Franck-Straße 1, 85748 Garching bei München');
            } else if (preg_match('/05\d{2}\.((EG)|\d{2})\.\d+/', $location)===1) {
                // TUM Campus Innenstadt
                self::switchLocation($event, $location, $matches[1].' Arcisstraße 21, 80333 München');
            } else if (preg_match('/01\d{2}\.((EG)|\d{2})\.\d+/', $location)===1) {
                // TUM Innenstadt Nordbau
                self::switchLocation($event, $location, $matches[1].' Theresienstraße 90, 80333 München');
            }
        }

        //Check status
        switch ($e->status) {
            default:
            case 'CONFIRMED':
                $event->setStatus(\Eluceo\iCal\Component\Event::STATUS_CONFIRMED);
                break;
            case 'CANCELLED':
                $event->setStatus(\Eluceo\iCal\Component\Event::STATUS_CANCELLED);
                break;
            case 'TENTATIVE':
                $event->setStatus(\Eluceo\iCal\Component\Event::STATUS_TENTATIVE);
                break;
        }

        //Add all fields
        $event->setUniqueId($e->uid)
            ->setDtStamp(new \DateTime($e->dtstamp))
            //->setUrl($e->)
            ->setSummary($summary)
            ->setDtStart(new \DateTime($e->dtstart))
            ->setDtEnd(new \DateTime($e->dtend));

        return $event;
    }


    /**
     * Update the location field
     *
     * @param $e array element to be edited
     * @param $newLoc string new location that should be set to the element
     */
    public static function switchLocation(\Eluceo\iCal\Component\Event &$e, $oldLocation, $newLoc) {
        $e->setDescription($oldLocation . "\n" . $e->getDescription());
        $e->setLocation($newLoc, $oldLocation);
    }

    /**
     * Remove duplicate entries: events that happen at the same time in multiple locations
     * @param $events
     */
    public static function noDupes(array &$events) {
        //Sort them
        usort($events, function (Event $a, Event $b) {
            if (strtotime($a->dtstart) > strtotime($b->dtstart)) {
                return 1;
            } else if ($a->dtstart > $b->dtstart) {
                return -1;
            }

            return 0;
        });

        //Find dupes
        $total = count($events);
        for ($i = 1; $i < $total; $i++) {
            //Check if start time, end time and title match then merge
            if ($events[$i - 1]->dtstart === $events[$i]->dtstart
                && $events[$i - 1]->dtend === $events[$i]->dtend
                && $events[$i - 1]->summary === $events[$i]->summary) {
                //Append the location to the next (same) element
                $events[$i]->location .= "\n" . $events[$i - 1]->location;

                //Mark this element for removal
                unset($events[$i - 1]);
            }
        }
    }
}
